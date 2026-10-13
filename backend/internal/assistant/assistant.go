// Package assistant implements the storefront AI perfume consultant: an
// OpenAI tool-calling loop that recommends perfumes from the catalog, edits
// the customer's cart, places orders and records leads. The cart lives in the
// browser, so cart edits are returned as actions for the frontend to apply;
// conversations are stored server-side per chat session.
//
// Every perfume is sold in volumes (variants): full bottles and decants
// ("raspiv"). Cart lines, totals and orders always refer to a variant id.
package assistant

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"

	"parfume/backend/internal/store"
)

const (
	maxHistory      = 20
	maxContentRunes = 2000
	maxToolRounds   = 8
	maxCatalogPages = 5 // × 100 products injected into the prompt
	maxQuantity     = 99
)

var (
	ErrDisabled       = errors.New("assistant is not configured")
	ErrNoPendingOrder = errors.New("no order is waiting for confirmation")

	sessionIDRe = regexp.MustCompile(`^[0-9a-f]{32}$`)
)

type Assistant struct {
	store  *store.Store
	client *openAIClient
}

// New returns nil when no API key is configured; a nil Assistant reports
// ErrDisabled from Chat.
func New(st *store.Store, apiKey, model string) *Assistant {
	if apiKey == "" {
		return nil
	}
	return &Assistant{store: st, client: newOpenAIClient(apiKey, model)}
}

// Request/response types shared with the HTTP layer.

type CartLine struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int   `json:"quantity"`
}

type Request struct {
	SessionID string     `json:"session_id"`
	Message   string     `json:"message"`
	Cart      []CartLine `json:"cart"`
	Locale    string     `json:"locale"`
}

// ShownProduct is a perfume card in the chat. VariantID is the volume the
// assistant suggests (the customer can switch volumes on the card).
type ShownProduct struct {
	Product   store.Product `json:"product"`
	VariantID int64         `json:"variant_id"`
	Quantity  int           `json:"quantity"`
	Note      string        `json:"note,omitempty"`
	Group     string        `json:"group,omitempty"`
}

type CartAction struct {
	Op        string         `json:"op"` // add | set | remove | clear
	VariantID int64          `json:"variant_id,omitempty"`
	Quantity  int            `json:"quantity,omitempty"`
	Product   *store.Product `json:"product,omitempty"`
}

type ComparisonRow struct {
	Label  string   `json:"label"`
	Values []string `json:"values"`
}

type Comparison struct {
	Products []store.Product `json:"products"`
	Rows     []ComparisonRow `json:"rows"`
	Verdict  string          `json:"verdict"`
}

type ProposalItem struct {
	Product  store.Product `json:"product"`
	Variant  store.Variant `json:"variant"`
	Quantity int           `json:"quantity"`
	Subtotal int64         `json:"subtotal"`
}

type OrderProposal struct {
	FullName string         `json:"full_name"`
	Phone    string         `json:"phone"`
	Comment  string         `json:"comment"`
	Items    []ProposalItem `json:"items"`
	Total    int64          `json:"total"`
}

type PlacedOrder struct {
	ID    int64 `json:"id"`
	Total int64 `json:"total"`
}

type Lead struct {
	Kind string `json:"kind"` // callback | quote
	ID   int64  `json:"id"`
}

type Reply struct {
	SessionID     string         `json:"session_id"`
	Reply         string         `json:"reply"`
	Products      []ShownProduct `json:"products"`
	CartActions   []CartAction   `json:"cart_actions"`
	OpenCart      bool           `json:"open_cart"`
	Comparison    *Comparison    `json:"comparison,omitempty"`
	OrderProposal *OrderProposal `json:"order_proposal,omitempty"`
	Order         *PlacedOrder   `json:"order,omitempty"`
	QuickReplies  []string       `json:"quick_replies"`
	Leads         []Lead         `json:"leads"`
}

// pendingOrder is what propose_order stores on the session until the
// customer confirms it.
type pendingOrder struct {
	FullName string     `json:"full_name"`
	Phone    string     `json:"phone"`
	Comment  string     `json:"comment"`
	Locale   string     `json:"locale"`
	Items    []CartLine `json:"items"`
}

// messageMeta is stored with each assistant message: product cards for the
// admin transcript, and notes that give the model context in later turns.
type messageMeta struct {
	Products []metaProduct `json:"products,omitempty"`
	Notes    []string      `json:"notes,omitempty"`
	OrderID  int64         `json:"order_id,omitempty"`
}

type metaProduct struct {
	ProductID int64  `json:"product_id"`
	VariantID int64  `json:"variant_id,omitempty"`
	Name      string `json:"name"`
	Volume    string `json:"volume,omitempty"`
	Quantity  int    `json:"quantity"`
	Group     string `json:"group,omitempty"`
}

// variantRef resolves a variant id to its perfume.
type variantRef struct {
	product store.Product
	variant store.Variant
}

// turn holds the state of a single /chat request while tools run.
type turn struct {
	session        store.ChatSession
	locale         string
	catalog        map[int64]store.Product
	variants       map[int64]variantRef
	cart           []CartLine // simulated cart, kept in sync with emitted actions
	pending        *pendingOrder
	pendingChanged bool
	notes          []string
	reply          Reply
}

func (t *turn) note(format string, args ...any) {
	t.notes = append(t.notes, fmt.Sprintf(format, args...))
}

func (t *turn) remember(p store.Product) {
	t.catalog[p.ID] = p
	for _, v := range p.Variants {
		t.variants[v.ID] = variantRef{product: p, variant: v}
	}
}

func (a *Assistant) Chat(ctx context.Context, req Request) (Reply, error) {
	if a == nil {
		return Reply{}, ErrDisabled
	}

	if !sessionIDRe.MatchString(req.SessionID) {
		req.SessionID = newSessionID()
	}
	session, err := a.store.EnsureChatSession(ctx, req.SessionID, req.Locale)
	if err != nil {
		return Reply{}, err
	}
	history, err := a.store.RecentChatMessages(ctx, session.ID, maxHistory)
	if err != nil {
		return Reply{}, err
	}
	products, err := a.loadCatalog(ctx)
	if err != nil {
		return Reply{}, err
	}
	settings, err := a.store.GetSettings(ctx)
	if err != nil {
		return Reply{}, err
	}

	t := &turn{
		session:  session,
		locale:   req.Locale,
		catalog:  make(map[int64]store.Product, len(products)),
		variants: map[int64]variantRef{},
		reply: Reply{
			SessionID:    session.ID,
			Products:     []ShownProduct{},
			CartActions:  []CartAction{},
			QuickReplies: []string{},
			Leads:        []Lead{},
		},
	}
	for _, p := range products {
		t.remember(p)
	}
	for _, line := range req.Cart {
		if _, ok := t.variants[line.VariantID]; ok && line.Quantity > 0 {
			t.cart = append(t.cart, CartLine{line.VariantID, min(line.Quantity, maxQuantity)})
		}
	}
	if len(session.PendingOrder) > 0 {
		var p pendingOrder
		if json.Unmarshal(session.PendingOrder, &p) == nil {
			t.pending = &p
		}
	}

	userContent := truncate(strings.TrimSpace(req.Message), maxContentRunes)
	messages := []chatMessage{
		{Role: "system", Content: ptr(systemPrompt(req.Locale, settings, products, t))},
	}
	messages = append(messages, historyMessages(history)...)
	messages = append(messages, chatMessage{Role: "user", Content: ptr(userContent)})

	if err := a.runLoop(ctx, t, messages); err != nil {
		return Reply{}, err
	}

	if t.pendingChanged {
		if err := a.store.SetPendingOrder(ctx, session.ID, t.pending); err != nil {
			log.Printf("assistant: save pending order: %v", err)
		}
	}
	a.saveTurn(ctx, t, userContent)
	return t.reply, nil
}

func (a *Assistant) runLoop(ctx context.Context, t *turn, messages []chatMessage) error {
	// Text the model writes alongside tool calls. It isn't replayed to the
	// model (which would make it repeat itself in the final answer); it's
	// only used when the final answer comes back empty.
	var draft string
	for range maxToolRounds {
		msg, err := a.client.complete(ctx, completionRequest{
			Messages:    messages,
			Tools:       tools,
			Temperature: 0.4,
			MaxTokens:   1500,
		})
		if err != nil {
			return err
		}
		if len(msg.ToolCalls) == 0 {
			text := draft
			if msg.Content != nil && strings.TrimSpace(*msg.Content) != "" {
				text = strings.TrimSpace(*msg.Content)
			}
			t.reply.Reply = extractTrailingOptions(t, extractLeakedReplies(t, text))
			return nil
		}

		if msg.Content != nil && strings.TrimSpace(*msg.Content) != "" {
			draft = strings.TrimSpace(*msg.Content)
		}
		msg.Content = nil
		messages = append(messages, msg)

		for _, call := range msg.ToolCalls {
			result := a.runTool(ctx, t, call.Function.Name, call.Function.Arguments)
			messages = append(messages, chatMessage{
				Role:       "tool",
				ToolCallID: call.ID,
				Content:    ptr(result),
			})
		}
	}
	t.reply.Reply = draft
	return nil
}

// saveTurn stores the exchange for later turns and the admin transcript.
// Failures are logged: the customer still gets the reply.
func (a *Assistant) saveTurn(ctx context.Context, t *turn, userContent string) {
	meta := messageMeta{Notes: t.notes}
	for _, s := range t.reply.Products {
		mp := metaProduct{
			ProductID: s.Product.ID, VariantID: s.VariantID, Name: fullName(s.Product),
			Quantity: s.Quantity, Group: s.Group,
		}
		if ref, ok := t.variants[s.VariantID]; ok {
			mp.Volume = volumeLabel(ref.variant)
		}
		meta.Products = append(meta.Products, mp)
	}
	if t.reply.Order != nil {
		meta.OrderID = t.reply.Order.ID
	}
	if err := a.store.AppendChatMessage(ctx, t.session.ID, "user", userContent, nil); err != nil {
		log.Printf("assistant: save message: %v", err)
		return
	}
	if err := a.store.AppendChatMessage(ctx, t.session.ID, "assistant", t.reply.Reply, meta); err != nil {
		log.Printf("assistant: save message: %v", err)
	}
}

// ConfirmOrder places the order the assistant proposed in this session.
func (a *Assistant) ConfirmOrder(ctx context.Context, sessionID string) (PlacedOrder, error) {
	if a == nil {
		return PlacedOrder{}, ErrDisabled
	}
	session, err := a.store.GetChatSession(ctx, sessionID)
	if err != nil {
		return PlacedOrder{}, err
	}
	var p pendingOrder
	if len(session.PendingOrder) == 0 || json.Unmarshal(session.PendingOrder, &p) != nil {
		return PlacedOrder{}, ErrNoPendingOrder
	}
	placed, err := a.placeOrder(ctx, sessionID, p)
	if err != nil {
		return PlacedOrder{}, err
	}
	note := fmt.Sprintf("Customer confirmed the order with the button. Order %s placed, total %d UZS.", OrderNumber(placed.ID), placed.Total)
	if err := a.store.AppendChatMessage(ctx, sessionID, "assistant", "", messageMeta{Notes: []string{note}, OrderID: placed.ID}); err != nil {
		log.Printf("assistant: save message: %v", err)
	}
	return placed, nil
}

func (a *Assistant) placeOrder(ctx context.Context, sessionID string, p pendingOrder) (PlacedOrder, error) {
	in := store.OrderInput{
		FullName:      p.FullName,
		Phone:         p.Phone,
		Comment:       p.Comment,
		Locale:        p.Locale,
		ChatSessionID: sessionID,
		Source:        "ai_chat",
	}
	for _, it := range p.Items {
		in.Items = append(in.Items, store.OrderItemInput{VariantID: it.VariantID, Quantity: it.Quantity})
	}
	order, err := a.store.CreateOrder(ctx, in)
	if err != nil {
		return PlacedOrder{}, err
	}
	return PlacedOrder{ID: order.ID, Total: order.Total}, nil
}

func (a *Assistant) loadCatalog(ctx context.Context) ([]store.Product, error) {
	var all []store.Product
	for page := 1; page <= maxCatalogPages; page++ {
		res, err := a.store.ListProducts(ctx, store.ProductFilter{ActiveOnly: true, Page: page, Limit: 100})
		if err != nil {
			return nil, err
		}
		all = append(all, res.Items...)
		if len(all) >= res.Total || len(res.Items) == 0 {
			break
		}
	}
	return all, nil
}

// historyMessages turns stored messages into model context, annotating
// assistant turns with the cards they showed and what their tools did, so
// follow-ups like "add the second one, but 10 ml" can be resolved.
func historyMessages(in []store.ChatMessage) []chatMessage {
	out := make([]chatMessage, 0, len(in))
	for _, m := range in {
		content := m.Content
		if m.Role == "assistant" {
			var meta messageMeta
			_ = json.Unmarshal(m.Meta, &meta)
			if len(meta.Products) > 0 {
				refs := make([]string, 0, len(meta.Products))
				for _, p := range meta.Products {
					ref := fmt.Sprintf("#%d %s", p.ProductID, p.Name)
					if p.VariantID > 0 {
						ref += fmt.Sprintf(" (suggested v%d %s ×%d)", p.VariantID, p.Volume, max(p.Quantity, 1))
					}
					if p.Group != "" {
						ref += " [" + p.Group + "]"
					}
					refs = append(refs, ref)
				}
				content += "\n[Perfume cards shown: " + strings.Join(refs, ", ") + "]"
			}
			for _, n := range meta.Notes {
				content += "\n[" + n + "]"
			}
		}
		content = strings.TrimSpace(content)
		if content == "" {
			continue
		}
		out = append(out, chatMessage{Role: m.Role, Content: ptr(content)})
	}
	return out
}

func systemPrompt(locale string, settings map[string]string, products []store.Product, t *turn) string {
	var b strings.Builder
	b.WriteString(`You are the AI perfume consultant of Parfume Shop — an online store of original perfumes in Uzbekistan. Every perfume is sold in volumes: full bottles and decants ("raspiv" in Uzbek, "распив" in Russian — 5-10 ml poured from the original bottle so customers can try a scent before buying a bottle). Prices are in Uzbek so'm (UZS). Your goal is to help each customer find a scent they will love and turn the conversation into an order or a consultant follow-up — honestly, without pressure.

LANGUAGE: Reply in the language the customer writes in (Uzbek Latin, Russian or English). If unclear, use the site language: ` + localeName(locale) + `. Notes, group names and quick replies must be in that language too.

UNDERSTAND THE TASTE (like a good perfume consultant):
- If the request is vague ("something nice", "a gift"), ask 1-2 short questions first and offer answer options with suggest_replies: for whom (her / him / unisex), occasion (every day / office / evening / special), season, preferred character (fresh, sweet, floral, woody, oriental/spicy), favourite perfumes they already know, and budget. Don't interrogate: if you know enough, recommend right away.
- Use the fragrance data: gender, concentration (EDP lasts longer than EDT; perfume oil is alcohol-free and close to the skin), family and the top/heart/base notes. Explain picks in plain words (e.g. "sweet vanilla and dates, warm, for cool evenings").
- If a customer names a perfume we don't sell, say so, call report_missing_product, and recommend the closest scents from our catalog by notes and character.

RECOMMENDING:
- Only recommend perfumes from the CATALOG (or that search_products returns). Never invent perfumes, notes, prices, volumes, discounts, stock or longevity claims beyond the data.
- Whenever you recommend specific perfumes, call show_products — the customer sees cards with volumes, prices and an "add to cart" button. Recommend 2-4 options, best match first, each with a very short note why it fits.
- In show_products, set variant_id to the volume you suggest: for customers who don't know the scent yet, suggest a decant to try first; for a gift or a known favourite, suggest a full bottle. Only use variants listed for that perfume and in stock.
- Cards already show names, volumes and prices, so keep your text to 1-3 short sentences — don't restate the cards as a list. End by asking whether to add one to the cart, with quick replies.

DECANTS ("raspiv"):
- Proactively suggest decants when the customer is unsure, wants to explore, or the bottle is expensive: "try 2-3 decants first, then buy the favourite as a bottle". A "discovery set" of decants is a great offer — group it with show_products "group".

BUDGET AND GIFTS:
- When the customer gives a budget, offer options that fit it; for gifts you may offer sets with the "group" field (e.g. "Sovg‘a uchun", "Sinab ko‘rish to‘plami"). Always verify sums with calculate_total — never do arithmetic yourself.

COMPARISON:
- When the customer hesitates between perfumes or asks for differences, call compare_products with 2-4 ids. Rows must only contain facts from the catalog data (notes, family, character from the description); write '—' where the data says nothing; never guess longevity or sillage. Brand, gender, concentration and prices are added automatically. Give a one-sentence verdict on who each suits.

CART (always by variant id = a specific volume):
- Call update_cart only when the customer clearly asks or confirms ("yes", "ha, qo‘sh", "добавь"...). Use the variant id of the volume they chose; if they didn't say which volume, use the volume suggested on the card.
- Resolve references like "the second one", "the 10 ml of Khamrah", "remove Sauvage" against the cards shown earlier and the CURRENT CART.
- "add" increases quantity, "set" sets the exact quantity, "remove" deletes the line. Use clear_first only if the customer asks to empty or replace the whole cart. Out-of-stock volumes can't be added.
- After changing the cart, confirm briefly what changed. Call open_cart when the customer wants to see the cart.

CHECKOUT IN CHAT:
- When the customer wants to order, make sure the cart has the right volumes, then ask for name and phone (and optional delivery address, gift wrapping, comment) if not given yet.
- Then call propose_order — the customer sees an order summary card with a "` + confirmLabel(locale) + `" button. Ask them to check it and press that button or confirm in chat.
- Call place_order ONLY if a proposed order is waiting and the customer explicitly confirms in their latest message ("ha, tasdiqlayman", "да, оформляйте", "yes, place it"). Never place an order on your own initiative. If the cart changes, propose again.
- After the order is placed, thank them, give the order number exactly as order_number (e.g. #00042) and say an operator will call to confirm delivery.

LEADS (consultant follow-up):
- request_callback: the customer wants a call, personal advice from a human, a perfume that is out of stock, or help ordering. Needs name and phone — ask for them, never invent them.
- request_quote: corporate gifts, wholesale/bulk orders or items we don't list. Include the volumes and quantities discussed. Needs name and phone (company optional).
- Only save a lead after the customer agreed to be contacted and gave their phone. Confirm afterwards that a consultant will call.

ORDER STATUS:
- To check an order, ask for the order number and the phone number used for it, then call check_order_status. Never reveal order details without a match.

QUICK REPLIES:
- Call the suggest_replies tool to offer 2-4 short tappable answers when a natural next step exists (e.g. "Ayollar uchun", "Shirin hidlar", "Raspiv qo‘shing", "Buyurtma berish"). Never write the options or the tool name into your text.

STYLE & SAFETY:
- Be warm, elegant, concise and practical. Plain text only: no markdown headings, tables, bold or links; short "- " bullet lists are fine.
- Don't promise discounts, delivery dates, originality certificates or prices that aren't in the data — offer a consultant follow-up instead. Skin reactions or allergies: suggest testing a decant on the wrist first.
- Politely decline topics unrelated to the store and perfumes.
`)

	b.WriteString("\nSTORE CONTACTS:\n")
	for _, k := range []string{"phone", "phone_secondary", "telegram_url", "address_uz", "work_hours_uz"} {
		if v := settings[k]; v != "" {
			fmt.Fprintf(&b, "- %s: %s\n", k, v)
		}
	}

	b.WriteString("\nCATALOG (#id | brand name | for | concentration | family | notes top / heart / base | volumes: v<variant id> ml kind price stock | description):\n")
	for _, p := range products {
		b.WriteString(catalogLine(p))
		b.WriteByte('\n')
	}
	if len(products) == maxCatalogPages*100 {
		b.WriteString("(Catalog truncated — use search_products to find other perfumes.)\n")
	}

	b.WriteString("\nCURRENT CART:\n")
	b.WriteString(t.cartSummary())

	b.WriteString("\n\nORDER WAITING FOR CONFIRMATION:\n")
	if t.pending != nil {
		fmt.Fprintf(&b, "Yes — for %s, %s (%d lines). place_order is allowed if the customer confirms now.", t.pending.FullName, t.pending.Phone, len(t.pending.Items))
	} else {
		b.WriteString("None.")
	}
	if t.session.OrderID != nil {
		fmt.Fprintf(&b, "\nThis customer already placed order %s in this chat.", OrderNumber(*t.session.OrderID))
	}

	b.WriteString(`

CHECKLIST BEFORE EVERY REPLY:
1. Did the customer ask for a perfume or brand we don't sell? → call report_missing_product once per such item (silently) and suggest similar scents from our catalog.
2. Are you recommending or mentioning specific catalog perfumes? → they must be shown with show_products in this same reply, each with a suggested variant_id. Never offer to show them "if you want" — show them now.
3. Never claim you contacted a consultant, saved a request, added to cart or placed an order unless the matching tool succeeded in this turn. Before you have the customer's name and phone, *offer* a follow-up instead.
4. Did the customer give name + phone and agree to be contacted or to order? → call the matching tool (request_callback / request_quote / propose_order) now.
5. Is there a natural next step? → call suggest_replies. Never write tool names, variant ids or JSON in your text.`)
	return b.String()
}

func catalogLine(p store.Product) string {
	family := "-"
	if p.Category != nil {
		family = p.Category.NameEn
	}
	volumes := make([]string, 0, len(p.Variants))
	for _, v := range p.Variants {
		stock := "in stock"
		if !v.InStock {
			stock = "OUT of stock"
		}
		volumes = append(volumes, fmt.Sprintf("v%d %dml %s %d %s", v.ID, v.VolumeML, v.Kind, v.Price, stock))
	}
	return fmt.Sprintf("#%d | %s | %s | %s | %s | %s / %s / %s | %s | %s",
		p.ID, fullName(p), p.Gender, p.Concentration, family,
		orDash(p.NotesTop), orDash(p.NotesHeart), orDash(p.NotesBase),
		strings.Join(volumes, "; "),
		truncate(firstNonEmpty(p.DescriptionEn, p.DescriptionUz), 200))
}

func (t *turn) cartSummary() string {
	if len(t.cart) == 0 {
		return "(empty)"
	}
	var b strings.Builder
	var total int64
	for _, line := range t.cart {
		ref := t.variants[line.VariantID]
		fmt.Fprintf(&b, "- v%d %s %s ×%d\n", line.VariantID, fullName(ref.product), volumeLabel(ref.variant), line.Quantity)
		total += ref.variant.Price * int64(line.Quantity)
	}
	fmt.Fprintf(&b, "Total: %d UZS", total)
	return b.String()
}

var leakedRepliesRe = regexp.MustCompile(`(?m)^\s*suggest_replies\b.*$`)

// extractLeakedReplies handles the model occasionally writing quick replies
// into its text instead of calling the tool, either as
// "suggest_replies: [...]" or as "suggest_replies:" followed by "- " lines.
func extractLeakedReplies(t *turn, text string) string {
	loc := leakedRepliesRe.FindStringIndex(text)
	if loc == nil {
		return text
	}
	var options []string
	header := text[loc[0]:loc[1]]
	rest := text[loc[1]:]
	if i := strings.Index(header, "["); i >= 0 {
		_ = json.Unmarshal([]byte(header[i:]), &options)
		rest = ""
	} else {
		lines := strings.Split(strings.TrimLeft(rest, "\n"), "\n")
		n := 0
		for ; n < len(lines); n++ {
			line := strings.TrimSpace(lines[n])
			if !strings.HasPrefix(line, "- ") {
				break
			}
			options = append(options, strings.TrimPrefix(line, "- "))
		}
		rest = strings.Join(lines[n:], "\n")
	}
	if len(t.reply.QuickReplies) == 0 {
		for _, o := range options {
			if o = strings.TrimSpace(o); o != "" && len(t.reply.QuickReplies) < 4 {
				t.reply.QuickReplies = append(t.reply.QuickReplies, truncate(o, 60))
			}
		}
	}
	return strings.TrimSpace(strings.TrimSpace(text[:loc[0]]) + "\n\n" + strings.TrimSpace(rest))
}

// extractTrailingOptions turns a short bullet list that ends the reply into
// quick replies when the model listed answer options instead of calling
// suggest_replies, e.g. "...qaysi biri?\n\n- Ayollar uchun\n- Erkaklar uchun".
func extractTrailingOptions(t *turn, text string) string {
	if len(t.reply.QuickReplies) > 0 {
		return text
	}
	lines := strings.Split(text, "\n")
	start := len(lines)
	for start > 0 {
		line := strings.TrimSpace(lines[start-1])
		if !strings.HasPrefix(line, "- ") || utf8.RuneCountInString(line) > 45 {
			break
		}
		start--
	}
	n := len(lines) - start
	if n < 2 || n > 4 || start == 0 {
		return text
	}
	// The options must follow a question, separated by a blank line.
	before := strings.TrimSpace(strings.Join(lines[:start], "\n"))
	if strings.TrimSpace(lines[start-1]) != "" || !(strings.HasSuffix(before, "?") || strings.HasSuffix(before, ":")) {
		return text
	}
	for _, line := range lines[start:] {
		t.reply.QuickReplies = append(t.reply.QuickReplies, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")))
	}
	return before
}

func newSessionID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func confirmLabel(locale string) string {
	switch locale {
	case "ru":
		return "Подтвердить заказ"
	case "en":
		return "Confirm order"
	}
	return "Buyurtmani tasdiqlash"
}

// OrderNumber formats an order id the way the storefront shows it.
func OrderNumber(id int64) string {
	return fmt.Sprintf("#%05d", id)
}

func localeName(locale string) string {
	switch locale {
	case "ru":
		return "Russian"
	case "en":
		return "English"
	}
	return "Uzbek (Latin script)"
}

// fullName is "Dior Sauvage".
func fullName(p store.Product) string {
	return strings.TrimSpace(p.Brand + " " + p.NameUz)
}

// volumeLabel is "10 ml decant" / "100 ml bottle".
func volumeLabel(v store.Variant) string {
	return fmt.Sprintf("%d ml %s", v.VolumeML, v.Kind)
}

func truncate(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func ptr[T any](v T) *T { return &v }

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
