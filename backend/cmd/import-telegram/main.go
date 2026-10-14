// Command import-telegram fills the catalog from a Telegram Desktop HTML export
// of the shop's channel (messages.html + photos/).
//
// Every post with a "Narxlari"/"Narxi" price list becomes a product; the same
// scent in another quality tier (#premium, #standart, #luxcopy, #tester) is a
// separate product. When a scent was posted several times, the newest post wins.
// Photos are downscaled and copied to <UPLOAD_DIR>/telegram/.
//
//	go run ./cmd/import-telegram -dir ~/Desktop/ChatExport_Mariana\ parfumes          # dry run
//	go run ./cmd/import-telegram -dir ~/Desktop/ChatExport_Mariana\ parfumes -apply   # write
//
// With -apply the demo catalog and products from a previous import are
// replaced; products added by hand in the admin panel are kept.
package main

import (
	"context"
	"flag"
	"fmt"
	"html"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"parfume/backend/internal/config"
	"parfume/backend/internal/database"
	"parfume/backend/internal/store"
)

const uploadSubdir = "telegram"

type post struct {
	ID     int
	Photos []string
	Text   string
}

type item struct {
	Key           string
	Post          int
	Tier          string
	Name, Brand   string
	Gender        string
	Concentration string
	Country       string
	Top, Heart    string
	Base          string
	Photo         string // path inside the export
	PhotoPost     int
	Variants      []store.VariantInput
}

func main() {
	dir := flag.String("dir", "", "Telegram export folder (contains messages.html)")
	apply := flag.Bool("apply", false, "write to the database (default: only print the plan)")
	flag.Parse()
	if *dir == "" {
		log.Fatal("-dir is required")
	}

	posts, err := readPosts(filepath.Join(*dir, "messages.html"))
	if err != nil {
		log.Fatal(err)
	}
	items, skipped := buildItems(posts)
	printPlan(items, skipped)
	if !*apply {
		fmt.Println("\nDry run. Add -apply to import.")
		return
	}

	cfg := config.Load()
	ctx := context.Background()
	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	if err := database.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	categories := map[string]int64{}
	rows, err := pool.Query(ctx, `SELECT slug, id FROM categories`)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var slug string
		var id int64
		if err := rows.Scan(&slug, &id); err != nil {
			log.Fatal(err)
		}
		categories[slug] = id
	}
	rows.Close()

	outDir := filepath.Join(cfg.UploadDir, uploadSubdir)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	tag, err := pool.Exec(ctx, `DELETE FROM products
		WHERE image_url LIKE '/images/perfumes/%' OR image_url LIKE '/uploads/`+uploadSubdir+`/%'`)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\nRemoved %d demo/previously imported products.\n", tag.RowsAffected())

	st := store.New(pool)
	featured := 0
	for i, it := range items {
		imageURL := ""
		if it.Photo != "" {
			name := filepath.Base(it.Photo)
			if err := copyResized(filepath.Join(*dir, it.Photo), filepath.Join(outDir, name), 1200); err != nil {
				log.Printf("photo %s: %v", it.Photo, err)
			} else {
				imageURL = "/uploads/" + uploadSubdir + "/" + name
			}
		}
		var catID *int64
		if id, ok := categories[family(it)]; ok {
			catID = &id
		}
		isFeatured := featured < 8 && it.Tier == "premium" && imageURL != ""
		if isFeatured {
			featured++
		}
		uz, ru, en := descriptions(it)
		_, err := st.CreateProduct(ctx, store.ProductInput{
			CategoryID:    catID,
			NameUz:        it.Name,
			NameRu:        it.Name,
			NameEn:        it.Name,
			DescriptionUz: uz,
			DescriptionRu: ru,
			DescriptionEn: en,
			Brand:         it.Brand,
			Gender:        it.Gender,
			Concentration: it.Concentration,
			Tier:          it.Tier,
			NotesTop:      it.Top,
			NotesHeart:    it.Heart,
			NotesBase:     it.Base,
			ImageURL:      imageURL,
			IsActive:      true,
			IsFeatured:    isFeatured,
			SortOrder:     i,
			Variants:      it.Variants,
		})
		if err != nil {
			log.Fatalf("post #%d %s: %v", it.Post, it.Name, err)
		}
	}
	fmt.Printf("Imported %d products (%d featured).\n", len(items), featured)
}

// ---------------------------------------------------------------- parsing

var (
	reID     = regexp.MustCompile(`id="message(\d+)"`)
	rePhoto  = regexp.MustCompile(`href="(photos/photo_[^"]+?\.jpg)"`)
	reText   = regexp.MustCompile(`(?s)<div class="text">\s*(.*?)\s*</div>`)
	reTag    = regexp.MustCompile(`<[^>]+>`)
	rePrice  = regexp.MustCompile(`(\d+)\s*ml\s*[-–—]\s*(\d[\d\s.,]*\d)`)
	reHash   = regexp.MustCompile(`#(premium|standart|luxcopy|tester)\b`)
	reSpaces = regexp.MustCompile(`\s+`)
)

func readPosts(path string) ([]post, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(raw), `<div class="message default`)
	posts := make([]post, 0, len(parts))
	for _, part := range parts[1:] {
		m := reID.FindStringSubmatch(part)
		if m == nil {
			continue
		}
		p := post{}
		p.ID, _ = strconv.Atoi(m[1])
		for _, ph := range rePhoto.FindAllStringSubmatch(part, -1) {
			p.Photos = append(p.Photos, ph[1])
		}
		if t := reText.FindStringSubmatch(part); t != nil {
			s := strings.ReplaceAll(t[1], "<br>", "\n")
			p.Text = strings.TrimSpace(html.UnescapeString(reTag.ReplaceAllString(s, "")))
		}
		posts = append(posts, p)
	}
	return posts, nil
}

type skip struct {
	Post   int
	Reason string
}

func buildItems(posts []post) ([]item, []skip) {
	var parsed []item
	var skipped []skip
	claimed := map[int]bool{}
	for i, p := range posts {
		if !strings.Contains(p.Text, "Narx") {
			continue
		}
		it, reason := parsePost(p)
		if reason != "" {
			skipped = append(skipped, skip{p.ID, reason})
			continue
		}
		// Product posts are often a video with the text, followed (or
		// preceded) by a text-less photo post of the bottle.
		if len(p.Photos) > 0 {
			it.Photo, it.PhotoPost = p.Photos[0], p.ID
		} else {
			for _, j := range []int{i + 1, i - 1} {
				if j < 0 || j >= len(posts) {
					continue
				}
				n := posts[j]
				if n.Text == "" && len(n.Photos) > 0 && !claimed[n.ID] {
					it.Photo, it.PhotoPost = n.Photos[0], n.ID
					claimed[n.ID] = true
					break
				}
			}
		}
		parsed = append(parsed, it)
	}

	// Newest post per scent and tier wins; older posts fill in what it lacks.
	byKey := map[string]*item{}
	var order []string
	for i := len(parsed) - 1; i >= 0; i-- {
		it := parsed[i]
		cur, ok := byKey[it.Key]
		if !ok {
			c := it
			byKey[it.Key] = &c
			order = append(order, it.Key)
			continue
		}
		if cur.Photo == "" {
			cur.Photo, cur.PhotoPost = it.Photo, it.PhotoPost
		}
		fillNotes(cur, it)
	}
	// Notes and photos are shared between tiers of the same scent.
	byScent := map[string][]*item{}
	for _, k := range order {
		it := byKey[k]
		byScent[scentKey(it.Brand, it.Name)] = append(byScent[scentKey(it.Brand, it.Name)], it)
	}
	for _, group := range byScent {
		for _, a := range group {
			for _, b := range group {
				fillNotes(a, *b)
				if a.Photo == "" && b.Photo != "" {
					a.Photo, a.PhotoPost = b.Photo, b.PhotoPost
				}
				if a.Concentration == "" {
					a.Concentration = b.Concentration
				}
			}
		}
	}

	out := make([]item, 0, len(order))
	for _, k := range order {
		it := byKey[k]
		if it.Concentration == "" {
			it.Concentration = "edp"
		}
		out = append(out, *it)
	}
	// Premium first, then standart, lux copy and testers; newest first inside a tier.
	rank := map[string]int{"premium": 0, "standart": 1, "luxcopy": 2, "tester": 3}
	sort.SliceStable(out, func(a, b int) bool {
		if rank[out[a].Tier] != rank[out[b].Tier] {
			return rank[out[a].Tier] < rank[out[b].Tier]
		}
		return out[a].Post > out[b].Post
	})
	return out, skipped
}

func fillNotes(dst *item, src item) {
	if dst.Top == "" && dst.Heart == "" && dst.Base == "" {
		dst.Top, dst.Heart, dst.Base = src.Top, src.Heart, src.Base
	}
	if dst.Country == "" {
		dst.Country = src.Country
	}
}

func parsePost(p post) (item, string) {
	it := item{Post: p.ID}
	m := reHash.FindStringSubmatch(p.Text)
	if m == nil {
		return it, "no tier hashtag"
	}
	it.Tier = m[1]

	var title string
	for _, line := range strings.Split(p.Text, "\n") {
		line = strings.TrimSpace(line)
		label, value, hasLabel := strings.Cut(line, ":")
		lower := strings.ToLower(label)
		switch {
		case hasLabel && strings.Contains(lower, "jins"):
			v := strings.ToLower(value)
			switch {
			case strings.Contains(v, "ayol"):
				it.Gender = "women"
			case strings.Contains(v, "erkak"):
				it.Gender = "men"
			default:
				it.Gender = "unisex"
			}
		case hasLabel && strings.Contains(lower, "brand"):
			it.Brand = cleanBrand(value)
		case hasLabel && strings.Contains(lower, "davlat"):
			it.Country = cleanWords(value)
		case hasLabel && strings.Contains(lower, "yuqori notalar"):
			it.Top = cleanNotes(value)
		case hasLabel && strings.Contains(lower, "rta notalar"):
			it.Heart = cleanNotes(value)
		case hasLabel && strings.Contains(lower, "asosiy notalar"):
			it.Base = cleanNotes(value)
		case strings.Contains(line, "《"):
			it.Concentration = concentration(line)
		case title == "" && line != "" && !strings.HasPrefix(line, "#"):
			title = line
		}
	}
	for _, pm := range rePrice.FindAllStringSubmatch(p.Text, -1) {
		ml, _ := strconv.Atoi(pm[1])
		digits := strings.Map(func(r rune) rune {
			if unicode.IsDigit(r) {
				return r
			}
			return -1
		}, pm[2])
		price, _ := strconv.ParseInt(digits, 10, 64)
		if ml <= 0 || ml > 1000 || price <= 0 {
			continue
		}
		kind := "bottle"
		if ml <= 10 && it.Tier != "tester" {
			kind = "decant"
		}
		it.Variants = append(it.Variants, store.VariantInput{VolumeML: ml, Kind: kind, Price: price, InStock: true})
	}
	if len(it.Variants) == 0 {
		return it, "no prices"
	}
	if it.Gender == "" {
		it.Gender = "unisex"
	}

	it.Name = perfumeName(cleanWords(title), it.Brand)
	if it.Name == "" {
		// A post without a scent name ("👑" only): fall back to the brand.
		it.Name = it.Brand
	}
	if it.Name == "" {
		return it, "no perfume name"
	}
	it.Key = it.Tier + "|" + scentKey(it.Brand, it.Name)
	return it, ""
}

func concentration(line string) string {
	l := strings.ToLower(line)
	switch {
	case strings.Contains(l, "extr") || strings.Contains(l, "extract"):
		return "extrait"
	case strings.Contains(l, "toilette") || strings.Contains(l, "edt"):
		return "edt"
	case strings.Contains(l, "cologne"):
		return "cologne"
	case strings.Contains(l, "eau de parfum") || strings.Contains(l, "edp"):
		return "edp"
	case strings.Contains(l, "parfum"):
		return "parfum"
	}
	return ""
}

// cleanWords drops emoji and decoration, keeping letters, digits and name punctuation.
func cleanWords(s string) string {
	s = strings.Map(func(r rune) rune {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r), unicode.IsSpace(r):
			return r
		case strings.ContainsRune("'’‘`ʼ&-.,()", r):
			if r == '’' || r == '‘' || r == '`' || r == 'ʼ' {
				return '\''
			}
			return r
		}
		return ' '
	}, s)
	return strings.Trim(reSpaces.ReplaceAllString(s, " "), " ,.-")
}

func cleanNotes(s string) string {
	parts := strings.Split(cleanWords(s), ",")
	out := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, ", ")
}

// Spelling fixes for brands as typed in the channel.
var brandFixes = map[string]string{
	"maison francis kurdjian": "Maison Francis Kurkdjian",
	"amouge":                  "Amouage",
	"yves saint laurant":      "Yves Saint Laurent",
	"ysl":                     "Yves Saint Laurent",
	"parfum de marly":         "Parfums de Marly",
	"creed":                   "Creed",
	"hermes":                  "Hermès",
	"ex nihilo":               "Ex Nihilo",
	"exnihilo":                "Ex Nihilo",
	"orto parisi":             "Orto Parisi",
	"le labo":                 "Le Labo",
	"marc-antoine barrois":    "Marc-Antoine Barrois",
	"marc - antoine barrois":  "Marc-Antoine Barrois",
	"marc antoine barrois":    "Marc-Antoine Barrois",
	"dolce & gabbana":         "Dolce & Gabbana",
	"dolce gabbana":           "Dolce & Gabbana",
	"victoria's secret":       "Victoria's Secret",
	"hugo boss":               "Hugo Boss",
	"gritti":                  "Gritti",
	"kilian":                  "Kilian",
	"hfc":                     "HFC",
	"roja":                    "Roja Parfums",
	"roja dove":               "Roja Parfums",
	"dima bilan":              "Dima Bilan",
	"by kilian":               "Kilian",
}

func cleanBrand(s string) string {
	b := cleanWords(s)
	if fixed, ok := brandFixes[strings.ToLower(b)]; ok {
		return fixed
	}
	if b == strings.ToUpper(b) {
		b = titleCase(b)
	}
	return b
}

// Perfume names as typed in the channel → canonical spelling (keys are lower-case, brand removed).
var nameFixes = map[string]string{
	"l'limmensite":                          "L'Immensité",
	"l'immensite":                           "L'Immensité",
	"llimmensite":                           "L'Immensité",
	"chanel de bleu":                        "Bleu de Chanel",
	"de bleu":                               "Bleu de Chanel",
	"bleu de chanel":                        "Bleu de Chanel",
	"bleu de":                               "Bleu de Chanel",
	"tygar":                                 "Tygar",
	"kirke":                                 "Kirke",
	"apres i'amour":                         "Après l'Amour",
	"apres l'amour":                         "Après l'Amour",
	"gucci flora gorgeus gardenia":          "Flora Gorgeous Gardenia",
	"flora gorgeus gardenia":                "Flora Gorgeous Gardenia",
	"flora gardenia":                        "Flora Gorgeous Gardenia",
	"flora gorgeous gardenia":               "Flora Gorgeous Gardenia",
	"my way eau de parfum":                  "My Way",
	"my way edp":                            "My Way",
	"libre edp":                             "Libre",
	"tutu":                                  "Tutu",
	"tutu blanc":                            "Tutu Blanc",
	"eau d'ombre leather":                   "Ombré Leather",
	"ombre leather":                         "Ombré Leather",
	"elysium":                               "Elysium",
	"roja elysium":                          "Elysium",
	"miss blooming bouquet":                 "Miss Dior Blooming Bouquet",
	"blue talisman extrait de parfum":       "Blue Talisman",
	"essential sport":                       "Essential Sport",
	"afternoon swim":                        "Afternoon Swim",
	"myslf":                                 "MYSLF",
	"excite by":                             "Excite",
	"02 molecules":                          "Escentric 02",
	"escentric 02 molecules":                "Escentric 02",
	"escentric 02":                          "Escentric 02",
	"baccarat rouge 540":                    "Baccarat Rouge 540",
	"mojave ghost":                          "Mojave Ghost",
	"wild vetiver":                          "Wild Vetiver",
	"terre d'hermes":                        "Terre d'Hermès",
	"angel's share by":                      "Angel's Share",
	"angel's share":                         "Angel's Share",
	"gaba hormone":                          "Gaba",
	"l'imperatrice":                         "L'Impératrice",
	"l'limperatrice":                        "L'Impératrice",
	"jump up and kiss me hedonistic (2021)": "Jump Up and Kiss Me Hedonistic",
}

// perfumeName removes the brand from a post title ("Imagination Louis Vuitton" → "Imagination").
func perfumeName(title, brand string) string {
	name := title
	if brand != "" {
		variants := []string{brand}
		for typed, fixed := range brandFixes {
			if fixed == brand {
				variants = append(variants, typed)
			}
		}
		// Longest first so "Marc - Antoine Barrois" goes before "Barrois".
		sort.Slice(variants, func(a, b int) bool { return len(variants[a]) > len(variants[b]) })
		for _, v := range variants {
			re := regexp.MustCompile(`(?i)(^|\s)` + regexp.QuoteMeta(v) + `($|\s)`)
			if stripped := strings.TrimSpace(re.ReplaceAllString(name, " ")); stripped != "" {
				name = stripped
			}
		}
	}
	name = strings.Trim(reSpaces.ReplaceAllString(name, " "), " ,.-")
	if fixed, ok := nameFixes[strings.ToLower(name)]; ok {
		return fixed
	}
	if name == strings.ToUpper(name) {
		name = titleCase(name)
	}
	return name
}

func titleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func scentKey(brand, name string) string {
	norm := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				return unicode.ToLower(r)
			}
			return -1
		}, s)
	}
	return norm(brand) + "/" + norm(name)
}

// ---------------------------------------------------------------- catalog

// family picks the fragrance family (category slug) that best matches the notes.
func family(it item) string {
	keywords := map[string][]string{
		"shirin":   {"vanil", "karamel", "pralin", "shokolad", "asal", "tonka", "olcha", "gilos", "kofe", "cherry", "malina", "qulupnay", "bodom", "sut"},
		"sharqona": {"ud", "oud", "shafran", "za'faron", "ambra", "tabak", "tamaki", "dolchin", "labdanum", "benzoin", "ladan", "smola", "ziravor", "koja", "teri", "charm"},
		"yogochli": {"yog'och", "yogoch", "kedr", "sandal", "vetiver", "pachuli", "gvayak", "moh", "archa", "kashmir", "papirus", "nagarmota"},
		"yangi":    {"sitrus", "sitron", "bergamot", "limon", "greypfruit", "greyfruit", "apelsin", "mandarin", "suv", "dengiz", "laym", "zanjabil", "imbir", "yalpiz", "lavanda", "choy", "ananas", "olma", "ambroksan", "ambroxan"},
		"gulli":    {"gul", "atirgul", "jasmin", "yasmin", "pion", "magnoliya", "iris", "binafsha", "fialka", "lotus", "tuberoza", "gardeniya", "nerol", "freziya", "lili", "osmantus", "geran", "gibiskus"},
	}
	weights := []struct {
		notes  string
		weight float64
	}{{it.Top, 1}, {it.Heart, 1.5}, {it.Base, 1.5}}
	scores := map[string]float64{}
	for _, w := range weights {
		for _, note := range strings.Split(strings.ToLower(w.notes), ",") {
			note = strings.TrimSpace(note)
			for fam, keys := range keywords {
				for _, k := range keys {
					if note == k || strings.HasPrefix(note, k) || strings.Contains(note, " "+k) {
						scores[fam] += w.weight
						break
					}
				}
			}
		}
	}
	best, bestScore := "", 0.0
	for _, fam := range []string{"sharqona", "shirin", "yogochli", "gulli", "yangi"} {
		if scores[fam] > bestScore {
			best, bestScore = fam, scores[fam]
		}
	}
	return best
}

func descriptions(it item) (uz, ru, en string) {
	tierUz := map[string]string{
		"premium":  "Premium sifat — eng yuqori daraja: hidi aslidek, turg‘unligi kuchli.",
		"standart": "Standart sifat — kundalik foydalanish uchun hamyonbop variant.",
		"luxcopy":  "Lux copy — sovg‘a uchun qulay: flakon va qadoq aslidek ishlangan.",
		"tester":   "Tester — qadoqsiz flakon, narxi qulayroq.",
	}
	tierRu := map[string]string{
		"premium":  "Премиум-качество — высший уровень: звучание как у оригинала, сильная стойкость.",
		"standart": "Стандарт — доступный вариант на каждый день.",
		"luxcopy":  "Люкс-копия — удобно для подарка: флакон и упаковка как у оригинала.",
		"tester":   "Тестер — флакон без коробки, по более выгодной цене.",
	}
	tierEn := map[string]string{
		"premium":  "Premium quality — our top tier: true-to-the-original scent with strong longevity.",
		"standart": "Standard quality — an affordable everyday option.",
		"luxcopy":  "Lux copy — a great gift: bottle and packaging modelled on the original.",
		"tester":   "Tester — unboxed bottle at a lower price.",
	}
	uz, ru, en = tierUz[it.Tier], tierRu[it.Tier], tierEn[it.Tier]
	if it.Country != "" {
		uz += " Brend mamlakati: " + it.Country + "."
	}
	return uz, ru, en
}

// ---------------------------------------------------------------- photos

// copyResized writes src as a JPEG no larger than max px on its long side.
func copyResized(src, dst string, max int) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w > max || h > max {
		if w >= h {
			w, h = max, h*max/w
		} else {
			w, h = w*max/h, max
		}
		img = downscale(img, w, h)
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 85}); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

// downscale averages the source pixels that fall into each target pixel (box filter).
func downscale(src image.Image, w, h int) image.Image {
	b := src.Bounds()
	sw, sh := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		y0, y1 := b.Min.Y+y*sh/h, b.Min.Y+(y+1)*sh/h
		for x := 0; x < w; x++ {
			x0, x1 := b.Min.X+x*sw/w, b.Min.X+(x+1)*sw/w
			var r, g, bl, n uint64
			for sy := y0; sy < y1; sy++ {
				for sx := x0; sx < x1; sx++ {
					cr, cg, cb, _ := src.At(sx, sy).RGBA()
					r, g, bl, n = r+uint64(cr), g+uint64(cg), bl+uint64(cb), n+1
				}
			}
			if n == 0 {
				continue
			}
			i := dst.PixOffset(x, y)
			dst.Pix[i+0] = uint8(r / n >> 8)
			dst.Pix[i+1] = uint8(g / n >> 8)
			dst.Pix[i+2] = uint8(bl / n >> 8)
			dst.Pix[i+3] = 0xff
		}
	}
	return dst
}

// ---------------------------------------------------------------- report

func printPlan(items []item, skipped []skip) {
	tiers := map[string]int{}
	noPhoto := 0
	for _, it := range items {
		tiers[it.Tier]++
		if it.Photo == "" {
			noPhoto++
		}
		var vs []string
		for _, v := range it.Variants {
			vs = append(vs, fmt.Sprintf("%d:%d", v.VolumeML, v.Price))
		}
		fmt.Printf("#%-4d %-8s %-28.28s %-38.38s %-6s %-7s %-8s photo#%-4d %s\n",
			it.Post, it.Tier, it.Brand, it.Name, it.Gender, it.Concentration, family(it), it.PhotoPost, strings.Join(vs, " "))
	}
	fmt.Printf("\n%d products: %d premium, %d standart, %d lux copy, %d tester; %d without photo\n",
		len(items), tiers["premium"], tiers["standart"], tiers["luxcopy"], tiers["tester"], noPhoto)
	for _, s := range skipped {
		fmt.Printf("skipped post #%d: %s\n", s.Post, s.Reason)
	}
}
