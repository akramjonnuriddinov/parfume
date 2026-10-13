package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"unicode/utf8"

	"parfume/backend/internal/phone"
	"parfume/backend/internal/store"
)

func object(props map[string]any, required ...string) map[string]any {
	if required == nil {
		required = []string{}
	}
	return map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             required,
		"additionalProperties": false,
	}
}

func str(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

var (
	productIDSchema = map[string]any{"type": "integer", "description": "Catalog perfume id (#id)"}
	variantIDSchema = map[string]any{"type": "integer", "description": "Variant id of a specific volume (v<id> in the catalog)"}
	itemsSchema     = map[string]any{
		"type": "array",
		"items": object(map[string]any{
			"variant_id": variantIDSchema,
			"quantity":   map[string]any{"type": "integer", "minimum": 1},
		}, "variant_id", "quantity"),
	}
	contactSchema = map[string]any{
		"full_name": str("Customer's name exactly as they gave it"),
		"phone":     str("Customer's phone number exactly as they gave it"),
	}
)

func withContact(props map[string]any) map[string]any {
	for k, v := range contactSchema {
		props[k] = v
	}
	return props
}

var tools = []toolDef{
	{Type: "function", Function: functionDef{
		Name:        "search_products",
		Description: "Search the catalog by perfume name, brand or note (e.g. 'vanilla', 'oud', 'Lattafa'). Use it when the item is not in the CATALOG list.",
		Parameters: object(map[string]any{
			"query": str("Short keyword"),
		}, "query"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "show_products",
		Description: "Show perfume cards (image, volumes, prices, add-to-cart) to the customer. Call it every time you recommend specific perfumes. Use 'group' to present alternative sets (e.g. a decant discovery set vs. a gift bottle).",
		Parameters: object(map[string]any{
			"items": map[string]any{
				"type":     "array",
				"minItems": 1,
				"maxItems": 12,
				"items": object(map[string]any{
					"product_id": productIDSchema,
					"variant_id": map[string]any{"type": "integer", "description": "Suggested volume: a variant id of this perfume (decant to try, bottle for gifts/known favourites)"},
					"quantity":   map[string]any{"type": "integer", "minimum": 1, "description": "Suggested quantity, usually 1"},
					"note":       str("Very short reason this scent fits, in the customer's language. Empty if not needed."),
					"group":      str("Set name when offering alternative sets, otherwise empty"),
				}, "product_id", "variant_id", "quantity", "note", "group"),
			},
		}, "items"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "compare_products",
		Description: "Show a side-by-side comparison of 2-4 perfumes. Brand, gender, concentration, notes and price rows are added automatically — add only other facts from the catalog data.",
		Parameters: object(map[string]any{
			"product_ids": map[string]any{"type": "array", "minItems": 2, "maxItems": 4, "items": productIDSchema},
			"rows": map[string]any{
				"type":     "array",
				"maxItems": 6,
				"items": object(map[string]any{
					"label":  str("Feature name in the customer's language, e.g. character, best season"),
					"values": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "One short value per perfume, same order as product_ids, in the customer's language; '—' if the data doesn't say"},
				}, "label", "values"),
			},
			"verdict": str("One sentence: who each option suits best"),
		}, "product_ids", "rows", "verdict"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "calculate_total",
		Description: "Calculate exact subtotals and the total for a list of volumes (variant ids) and quantities. Use it for budgets — never sum prices yourself.",
		Parameters:  object(map[string]any{"items": itemsSchema}, "items"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "update_cart",
		Description: "Change the customer's cart (lines are specific volumes). Only call after the customer asked or confirmed.",
		Parameters: object(map[string]any{
			"clear_first": map[string]any{"type": "boolean", "description": "Empty the cart first. Only when the customer explicitly asks to empty or replace the whole cart."},
			"operations": map[string]any{
				"type": "array",
				"items": object(map[string]any{
					"action":     map[string]any{"type": "string", "enum": []string{"add", "set", "remove"}},
					"variant_id": variantIDSchema,
					"quantity":   map[string]any{"type": "integer", "minimum": 0, "description": "For add: how many to add. For set: exact quantity. Ignored for remove."},
				}, "action", "variant_id", "quantity"),
			},
		}, "operations"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "open_cart",
		Description: "Open the cart panel so the customer can review it.",
		Parameters:  object(map[string]any{}),
	}},
	{Type: "function", Function: functionDef{
		Name:        "propose_order",
		Description: "Prepare an order from the CURRENT CART and show the customer a summary card with a 'Confirm order' button. Requires the customer's name and phone.",
		Parameters: object(withContact(map[string]any{
			"comment": str("Delivery address, gift wrapping or other wishes the customer mentioned; empty if none"),
		}), "full_name", "phone", "comment"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "place_order",
		Description: "Place the proposed order. ONLY when an order is waiting for confirmation and the customer explicitly confirmed it in their latest message.",
		Parameters:  object(map[string]any{}),
	}},
	{Type: "function", Function: functionDef{
		Name:        "request_callback",
		Description: "Save a request for a consultant to call the customer back (personal advice, out-of-stock perfume, help with ordering).",
		Parameters: object(withContact(map[string]any{
			"product_id": map[string]any{"type": "integer", "description": "Perfume the request is about, or 0"},
			"message":    str("Short summary of what the customer needs, for the consultant"),
		}), "full_name", "phone", "product_id", "message"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "request_quote",
		Description: "Save a request for corporate gifts, wholesale/bulk orders or perfumes we don't list, for a manager.",
		Parameters: object(withContact(map[string]any{
			"organization": str("Company name, or empty"),
			"items":        itemsSchema,
			"comment":      str("Other requirements, incl. perfumes not in our catalog; empty if none"),
		}), "full_name", "phone", "organization", "items", "comment"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "check_order_status",
		Description: "Look up an order's status. Requires the order number and the phone number used for the order.",
		Parameters: object(map[string]any{
			"order_id": map[string]any{"type": "integer", "description": "Order number without '#' or leading zeros"},
			"phone":    str("Phone number the customer used for the order"),
		}, "order_id", "phone"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "report_missing_product",
		Description: "Internal demand statistics: record that the customer wanted a perfume or brand we don't sell. This does NOT notify anyone — never tell the customer a consultant was informed because of it.",
		Parameters: object(map[string]any{
			"name":    str("Perfume as 'Brand Name', e.g. 'Creed Aventus', or a brand"),
			"details": str("Context in a few words, e.g. 'gift for husband, 100 ml'"),
		}, "name", "details"),
	}},
	{Type: "function", Function: functionDef{
		Name:        "suggest_replies",
		Description: "Show 2-4 short tappable reply buttons for the customer's next message.",
		Parameters: object(map[string]any{
			"options": map[string]any{"type": "array", "minItems": 2, "maxItems": 4, "items": str("Up to 40 characters, in the customer's language")},
		}, "options"),
	}},
}

func (a *Assistant) runTool(ctx context.Context, t *turn, name, args string) string {
	var err error
	var result any
	switch name {
	case "search_products":
		result, err = a.searchProducts(ctx, t, args)
	case "show_products":
		result, err = a.showProducts(ctx, t, args)
	case "compare_products":
		result, err = a.compareProducts(ctx, t, args)
	case "calculate_total":
		result, err = a.calculateTotal(ctx, t, args)
	case "update_cart":
		result, err = a.updateCart(ctx, t, args)
	case "open_cart":
		t.reply.OpenCart = true
		result = map[string]bool{"ok": true}
	case "propose_order":
		result, err = a.proposeOrder(ctx, t, args)
	case "place_order":
		result, err = a.placeOrderTool(ctx, t)
	case "request_callback":
		result, err = a.requestCallback(ctx, t, args)
	case "request_quote":
		result, err = a.requestQuote(ctx, t, args)
	case "check_order_status":
		result, err = a.checkOrderStatus(ctx, args)
	case "report_missing_product":
		result, err = a.reportMissing(ctx, t, args)
	case "suggest_replies":
		result, err = suggestReplies(t, args)
	default:
		err = fmt.Errorf("unknown tool %q", name)
	}
	if err != nil {
		return mustJSON(map[string]string{"error": err.Error()})
	}
	return mustJSON(result)
}

type itemRef struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int   `json:"quantity"`
}

func (a *Assistant) searchProducts(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	res, err := a.store.ListProducts(ctx, store.ProductFilter{ActiveOnly: true, Query: strings.TrimSpace(in.Query), Limit: 15})
	if err != nil {
		return nil, err
	}
	lines := make([]string, 0, len(res.Items))
	for _, p := range res.Items {
		t.remember(p)
		lines = append(lines, catalogLine(p))
	}
	return map[string]any{"results": lines}, nil
}

func (a *Assistant) showProducts(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		Items []struct {
			ProductID int64  `json:"product_id"`
			VariantID int64  `json:"variant_id"`
			Quantity  int    `json:"quantity"`
			Note      string `json:"note"`
			Group     string `json:"group"`
		} `json:"items"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	var shown, missing, fixed []string
	for _, it := range in.Items {
		p, ok := a.product(ctx, t, it.ProductID)
		if !ok {
			// The model sometimes passes a variant id as the product id.
			if ref, isVariant := t.variants[it.ProductID]; isVariant {
				p, ok, it.VariantID = ref.product, true, it.ProductID
			}
		}
		if !ok {
			missing = append(missing, fmt.Sprintf("#%d", it.ProductID))
			continue
		}
		variantID := suggestedVariant(p, it.VariantID)
		if variantID != it.VariantID {
			fixed = append(fixed, fmt.Sprintf("#%d → v%d", p.ID, variantID))
		}
		item := ShownProduct{
			Product:   p,
			VariantID: variantID,
			Quantity:  clampQty(it.Quantity),
			Note:      truncate(strings.TrimSpace(it.Note), 200),
			Group:     truncate(strings.TrimSpace(it.Group), 40),
		}
		// A later card for the same perfume in the same set replaces the earlier one.
		t.reply.Products = slices.DeleteFunc(t.reply.Products, func(s ShownProduct) bool {
			return s.Product.ID == p.ID && s.Group == item.Group
		})
		t.reply.Products = append(t.reply.Products, item)
		shown = append(shown, fmt.Sprintf("#%d", p.ID))
	}
	result := map[string]any{"shown": shown, "not_found": missing}
	if len(fixed) > 0 {
		result["volume_adjusted"] = fixed
	}
	return result, nil
}

// suggestedVariant keeps the requested volume when it belongs to the perfume
// and is in stock, otherwise falls back to the first in-stock one.
func suggestedVariant(p store.Product, want int64) int64 {
	var fallback int64
	for _, v := range p.Variants {
		if v.ID == want && v.InStock {
			return v.ID
		}
		if fallback == 0 && v.InStock {
			fallback = v.ID
		}
	}
	return fallback
}

func (a *Assistant) compareProducts(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		ProductIDs []int64         `json:"product_ids"`
		Rows       []ComparisonRow `json:"rows"`
		Verdict    string          `json:"verdict"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	c := &Comparison{Rows: []ComparisonRow{}, Verdict: truncate(strings.TrimSpace(in.Verdict), 300)}
	for _, id := range in.ProductIDs {
		if p, ok := a.product(ctx, t, id); ok && len(c.Products) < 4 {
			c.Products = append(c.Products, p)
		}
	}
	if len(c.Products) < 2 {
		return nil, errors.New("need at least 2 existing perfumes")
	}
	for _, row := range in.Rows {
		if len(c.Rows) == 6 {
			break
		}
		values := make([]string, len(c.Products))
		for i := range values {
			values[i] = "—"
			if i < len(row.Values) && strings.TrimSpace(row.Values[i]) != "" {
				values[i] = truncate(strings.TrimSpace(row.Values[i]), 80)
			}
		}
		c.Rows = append(c.Rows, ComparisonRow{Label: truncate(strings.TrimSpace(row.Label), 40), Values: values})
	}
	t.reply.Comparison = c
	ids := make([]string, len(c.Products))
	for i, p := range c.Products {
		ids[i] = fmt.Sprintf("#%d", p.ID)
	}
	t.note("Comparison table shown: %s", strings.Join(ids, " vs "))
	return map[string]any{"shown": ids}, nil
}

func (a *Assistant) calculateTotal(_ context.Context, t *turn, args string) (any, error) {
	var in struct {
		Items []itemRef `json:"items"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	type line struct {
		VariantID int64  `json:"variant_id"`
		Quantity  int    `json:"quantity"`
		Subtotal  *int64 `json:"subtotal"`
		Note      string `json:"note,omitempty"`
	}
	var lines []line
	var total int64
	for _, it := range in.Items {
		l := line{VariantID: it.VariantID, Quantity: clampQty(it.Quantity)}
		ref, ok := t.variants[it.VariantID]
		switch {
		case !ok:
			l.Note = "variant not found"
		case !ref.variant.InStock:
			l.Note = "out of stock — not included"
		default:
			sub := ref.variant.Price * int64(l.Quantity)
			l.Subtotal = &sub
			total += sub
		}
		lines = append(lines, l)
	}
	return map[string]any{"lines": lines, "total_uzs": total}, nil
}

func (a *Assistant) updateCart(_ context.Context, t *turn, args string) (any, error) {
	var in struct {
		ClearFirst bool `json:"clear_first"`
		Operations []struct {
			Action    string `json:"action"`
			VariantID int64  `json:"variant_id"`
			Quantity  int    `json:"quantity"`
		} `json:"operations"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	if in.ClearFirst {
		t.cart = nil
		t.reply.CartActions = append(t.reply.CartActions, CartAction{Op: "clear"})
	}
	var errs []string
	for _, op := range in.Operations {
		if err := t.applyCartOp(op.Action, op.VariantID, op.Quantity); err != "" {
			errs = append(errs, err)
		}
	}
	// A proposed order no longer matches the cart.
	if t.pending != nil {
		t.pending = nil
		t.pendingChanged = true
		t.reply.OrderProposal = nil
	}
	return map[string]any{"errors": errs, "cart_now": t.cartSummary()}, nil
}

func (t *turn) applyCartOp(action string, id int64, qty int) string {
	idx := slices.IndexFunc(t.cart, func(l CartLine) bool { return l.VariantID == id })

	if action == "remove" || (action == "set" && qty <= 0) {
		if idx >= 0 {
			t.cart = slices.Delete(t.cart, idx, idx+1)
		}
		t.reply.CartActions = append(t.reply.CartActions, CartAction{Op: "remove", VariantID: id})
		return ""
	}
	if action != "add" && action != "set" {
		return fmt.Sprintf("v%d: unknown action %q", id, action)
	}

	ref, ok := t.variants[id]
	switch {
	case !ok:
		return fmt.Sprintf("v%d: no such volume — use a variant id (v<id>) from the catalog", id)
	case !ref.variant.InStock:
		return fmt.Sprintf("v%d: this volume is out of stock — offer another volume or a callback", id)
	}

	qty = clampQty(qty)
	switch {
	case idx < 0:
		t.cart = append(t.cart, CartLine{VariantID: id, Quantity: qty})
	case action == "add":
		t.cart[idx].Quantity = min(t.cart[idx].Quantity+qty, maxQuantity)
	default:
		t.cart[idx].Quantity = qty
	}
	p := ref.product
	t.reply.CartActions = append(t.reply.CartActions, CartAction{Op: action, VariantID: id, Quantity: qty, Product: &p})
	return ""
}

func (a *Assistant) proposeOrder(_ context.Context, t *turn, args string) (any, error) {
	var in struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Comment  string `json:"comment"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	name, tel, err := validContact(in.FullName, in.Phone)
	if err != nil {
		return nil, err
	}
	comment := strings.TrimSpace(in.Comment)
	if utf8.RuneCountInString(comment) > 1000 {
		return nil, errors.New("comment is too long")
	}
	if len(t.cart) == 0 {
		return nil, errors.New("the cart is empty — add perfumes first")
	}
	if len(t.cart) > 50 {
		return nil, errors.New("too many lines for one order (max 50)")
	}

	proposal := &OrderProposal{FullName: name, Phone: tel, Comment: comment}
	pending := &pendingOrder{FullName: name, Phone: tel, Comment: comment, Locale: t.locale}
	for _, line := range t.cart {
		ref, ok := t.variants[line.VariantID]
		if !ok || !ref.variant.InStock {
			return nil, fmt.Errorf("v%d is not available — remove it from the cart first", line.VariantID)
		}
		sub := ref.variant.Price * int64(line.Quantity)
		proposal.Items = append(proposal.Items, ProposalItem{Product: ref.product, Variant: ref.variant, Quantity: line.Quantity, Subtotal: sub})
		proposal.Total += sub
		pending.Items = append(pending.Items, line)
	}

	t.pending = pending
	t.pendingChanged = true
	t.reply.OrderProposal = proposal
	t.note("Order summary shown for %s, %s, total %d UZS — waiting for the customer to confirm.", name, tel, proposal.Total)
	items := make([]string, len(proposal.Items))
	for i, it := range proposal.Items {
		items[i] = fmt.Sprintf("%s, %s ×%d", fullName(it.Product), volumeLabel(it.Variant), it.Quantity)
	}
	return map[string]any{
		"status":  "summary card shown; ask the customer to check it and press Confirm (or confirm in chat)",
		"items":   items,
		"note":    "These are the exact lines from the customer's cart — describe these, not earlier suggestions.",
		"total":   proposal.Total,
		"phone":   tel,
		"comment": comment,
	}, nil
}

func (a *Assistant) placeOrderTool(ctx context.Context, t *turn) (any, error) {
	if t.pending == nil {
		return nil, errors.New("no order is waiting for confirmation — call propose_order first")
	}
	placed, err := a.placeOrder(ctx, t.session.ID, *t.pending)
	if err != nil {
		if errors.Is(err, store.ErrValidation) {
			return nil, fmt.Errorf("could not place the order: %v", err)
		}
		return nil, err
	}
	t.pending = nil
	t.pendingChanged = false // CreateOrder already cleared it
	t.reply.OrderProposal = nil
	t.reply.Order = &placed
	t.note("Order %s placed, total %d UZS.", OrderNumber(placed.ID), placed.Total)
	return map[string]any{"order_number": OrderNumber(placed.ID), "total": placed.Total}, nil
}

func (a *Assistant) requestCallback(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		FullName  string `json:"full_name"`
		Phone     string `json:"phone"`
		ProductID int64  `json:"product_id"`
		Message   string `json:"message"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	name, tel, err := validContact(in.FullName, in.Phone)
	if err != nil {
		return nil, err
	}
	msg := store.MessageInput{
		FullName: name,
		Phone:    tel,
		Message:  truncate(strings.TrimSpace(in.Message), 1500),
		Source:   "ai_callback",
		ChatID:   t.session.ID,
	}
	if _, ok := a.product(ctx, t, in.ProductID); ok {
		msg.ProductID = &in.ProductID
	}
	return a.saveLead(ctx, t, "callback", msg)
}

func (a *Assistant) requestQuote(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		FullName     string    `json:"full_name"`
		Phone        string    `json:"phone"`
		Organization string    `json:"organization"`
		Items        []itemRef `json:"items"`
		Comment      string    `json:"comment"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	name, tel, err := validContact(in.FullName, in.Phone)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteString("Korporativ / ulgurji so‘rov (AI chat)\n")
	if org := strings.TrimSpace(in.Organization); org != "" {
		fmt.Fprintf(&b, "Kompaniya: %s\n", truncate(org, 200))
	}
	if len(in.Items) > 0 {
		b.WriteString("\n")
		var total int64
		for _, it := range in.Items[:min(len(in.Items), 50)] {
			ref, ok := t.variants[it.VariantID]
			if !ok {
				continue
			}
			qty := clampQty(it.Quantity)
			sub := ref.variant.Price * int64(qty)
			total += sub
			fmt.Fprintf(&b, "- %s, %d ml × %d — %s\n", fullName(ref.product), ref.variant.VolumeML, qty, money(sub))
		}
		fmt.Fprintf(&b, "Jami (katalog narxlarida): %s\n", money(total))
	}
	if c := strings.TrimSpace(in.Comment); c != "" {
		fmt.Fprintf(&b, "\nIzoh: %s\n", truncate(c, 1000))
	}

	return a.saveLead(ctx, t, "quote", store.MessageInput{
		FullName: name,
		Phone:    tel,
		Message:  strings.TrimSpace(b.String()),
		Source:   "ai_quote",
		ChatID:   t.session.ID,
	})
}

func (a *Assistant) saveLead(ctx context.Context, t *turn, kind string, in store.MessageInput) (any, error) {
	m, err := a.store.CreateMessage(ctx, in)
	if err != nil {
		return nil, err
	}
	if err := a.store.IncrementChatLeads(ctx, t.session.ID); err != nil {
		log.Printf("assistant: count lead: %v", err)
	}
	t.reply.Leads = append(t.reply.Leads, Lead{Kind: kind, ID: m.ID})
	t.note("%s request #%d saved for %s, %s — a consultant will call.", kind, m.ID, in.FullName, in.Phone)
	return map[string]any{"saved": true, "request_id": m.ID}, nil
}

func (a *Assistant) checkOrderStatus(ctx context.Context, args string) (any, error) {
	var in struct {
		OrderID json.Number `json:"order_id"`
		Phone   string      `json:"phone"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	notFound := map[string]any{"found": false, "hint": "No order matches this number and phone. Ask the customer to double-check both."}

	id, err := strconv.ParseInt(strings.TrimLeft(strings.TrimPrefix(in.OrderID.String(), "#"), "0"), 10, 64)
	tel, ok := phone.Normalize(in.Phone)
	if err != nil || !ok {
		return notFound, nil
	}
	o, err := a.store.GetOrder(ctx, id)
	if errors.Is(err, store.ErrNotFound) || (err == nil && o.Phone != tel) {
		return notFound, nil
	}
	if err != nil {
		return nil, err
	}

	items := make([]string, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, fmt.Sprintf("%s %s, %d ml × %d", it.ProductBrand, it.ProductName, it.VolumeML, it.Quantity))
	}
	return map[string]any{
		"found":        true,
		"order_number": OrderNumber(o.ID),
		"status":       o.Status,
		"status_meaning": map[string]string{
			"new":        "received, an operator will call to confirm",
			"processing": "confirmed and being prepared/delivered",
			"completed":  "delivered",
			"cancelled":  "cancelled",
		}[o.Status],
		"created_at": o.CreatedAt.Format("2006-01-02 15:04"),
		"total_uzs":  o.Total,
		"items":      items,
	}, nil
}

func (a *Assistant) reportMissing(ctx context.Context, t *turn, args string) (any, error) {
	var in struct {
		Name    string `json:"name"`
		Details string `json:"details"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	name := strings.ToLower(truncate(strings.TrimSpace(in.Name), 80))
	if name == "" {
		return nil, errors.New("name is required")
	}
	if err := a.store.AddMissingProduct(ctx, t.session.ID, name, truncate(strings.TrimSpace(in.Details), 300)); err != nil {
		return nil, err
	}
	t.note("Missing perfume logged for statistics: %s", name)
	return map[string]any{
		"recorded": true,
		"note":     "Internal statistics only — nobody was notified. Recommend similar scents from the catalog; to involve a consultant, ask for name and phone and use request_callback.",
	}, nil
}

func suggestReplies(t *turn, args string) (any, error) {
	var in struct {
		Options []string `json:"options"`
	}
	if err := json.Unmarshal([]byte(args), &in); err != nil {
		return nil, err
	}
	t.reply.QuickReplies = t.reply.QuickReplies[:0]
	for _, o := range in.Options {
		if o = strings.TrimSpace(o); o != "" && len(t.reply.QuickReplies) < 4 {
			t.reply.QuickReplies = append(t.reply.QuickReplies, truncate(o, 60))
		}
	}
	return map[string]bool{"shown": true}, nil
}

// product looks up an active perfume, falling back to the database for items
// beyond the catalog injected into the prompt.
func (a *Assistant) product(ctx context.Context, t *turn, id int64) (store.Product, bool) {
	if id <= 0 {
		return store.Product{}, false
	}
	if p, ok := t.catalog[id]; ok {
		return p, true
	}
	p, err := a.store.GetProduct(ctx, id)
	if err != nil || !p.IsActive {
		return store.Product{}, false
	}
	t.remember(p)
	return p, true
}

func validContact(name, rawPhone string) (string, string, error) {
	name = strings.TrimSpace(name)
	if n := utf8.RuneCountInString(name); n < 2 || n > 120 {
		return "", "", errors.New("ask the customer for their name (2-120 characters)")
	}
	tel, ok := phone.Normalize(rawPhone)
	if !ok {
		return "", "", errors.New("invalid phone number — ask the customer for a valid one, e.g. +998 90 123 45 67")
	}
	return name, tel, nil
}

func clampQty(q int) int {
	return max(1, min(q, maxQuantity))
}

// money formats 1250000 as "1 250 000 so‘m".
func money(v int64) string {
	s := strconv.FormatInt(v, 10)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String() + " so‘m"
}
