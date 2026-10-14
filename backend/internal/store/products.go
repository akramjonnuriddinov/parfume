package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	Genders        = []string{"women", "men", "unisex"}
	Concentrations = []string{"edp", "edt", "parfum", "extrait", "cologne", "oil"}
	VariantKinds   = []string{"bottle", "decant"}
	Tiers          = []string{"premium", "standart", "luxcopy", "tester"}
)

func oneOf(v string, list []string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func ValidGender(s string) bool        { return oneOf(s, Genders) }
func ValidConcentration(s string) bool { return oneOf(s, Concentrations) }
func ValidVariantKind(s string) bool   { return oneOf(s, VariantKinds) }
func ValidTier(s string) bool          { return oneOf(s, Tiers) }

type CategoryRef struct {
	ID     int64  `json:"id"`
	Slug   string `json:"slug"`
	NameUz string `json:"name_uz"`
	NameRu string `json:"name_ru"`
	NameEn string `json:"name_en"`
}

// Variant is one sellable volume of a perfume: a full bottle or a decant.
type Variant struct {
	ID        int64  `json:"id"`
	VolumeML  int    `json:"volume_ml"`
	Kind      string `json:"kind"`
	Price     int64  `json:"price"`
	InStock   bool   `json:"in_stock"`
	SortOrder int    `json:"sort_order"`
}

type Product struct {
	ID            int64        `json:"id"`
	CategoryID    *int64       `json:"category_id"`
	Category      *CategoryRef `json:"category"`
	NameUz        string       `json:"name_uz"`
	NameRu        string       `json:"name_ru"`
	NameEn        string       `json:"name_en"`
	DescriptionUz string       `json:"description_uz"`
	DescriptionRu string       `json:"description_ru"`
	DescriptionEn string       `json:"description_en"`
	Brand         string       `json:"brand"`
	Gender        string       `json:"gender"`
	Concentration string       `json:"concentration"`
	Tier          string       `json:"tier"`
	NotesTop      string       `json:"notes_top"`
	NotesHeart    string       `json:"notes_heart"`
	NotesBase     string       `json:"notes_base"`
	ImageURL      string       `json:"image_url"`
	IsActive      bool         `json:"is_active"`
	IsFeatured    bool         `json:"is_featured"`
	SortOrder     int          `json:"sort_order"`
	Variants      []Variant    `json:"variants"`
	// MinPrice is the cheapest in-stock volume ("from 45 000 so'm"); nil when nothing is in stock.
	MinPrice  *int64    `json:"min_price"`
	InStock   bool      `json:"in_stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type VariantInput struct {
	ID        int64  `json:"id"` // 0 = new
	VolumeML  int    `json:"volume_ml"`
	Kind      string `json:"kind"`
	Price     int64  `json:"price"`
	InStock   bool   `json:"in_stock"`
	SortOrder int    `json:"sort_order"`
}

type ProductInput struct {
	CategoryID    *int64         `json:"category_id"`
	NameUz        string         `json:"name_uz"`
	NameRu        string         `json:"name_ru"`
	NameEn        string         `json:"name_en"`
	DescriptionUz string         `json:"description_uz"`
	DescriptionRu string         `json:"description_ru"`
	DescriptionEn string         `json:"description_en"`
	Brand         string         `json:"brand"`
	Gender        string         `json:"gender"`
	Concentration string         `json:"concentration"`
	Tier          string         `json:"tier"`
	NotesTop      string         `json:"notes_top"`
	NotesHeart    string         `json:"notes_heart"`
	NotesBase     string         `json:"notes_base"`
	ImageURL      string         `json:"image_url"`
	IsActive      bool           `json:"is_active"`
	IsFeatured    bool           `json:"is_featured"`
	SortOrder     int            `json:"sort_order"`
	Variants      []VariantInput `json:"variants"`
}

type ProductFilter struct {
	ActiveOnly   bool
	CategoryID   int64
	CategorySlug string
	Gender       string
	Brand        string
	Tier         string
	Query        string
	Featured     bool
	IDs          []int64
	Sort         string
	Page         int
	Limit        int
}

const productSelect = `
	SELECT p.id, p.category_id, p.name_uz, p.name_ru, p.name_en,
		p.description_uz, p.description_ru, p.description_en, p.brand, p.gender, p.concentration, p.tier,
		p.notes_top, p.notes_heart, p.notes_base,
		p.image_url, p.is_active, p.is_featured, p.sort_order,
		p.created_at, p.updated_at,
		c.id, c.slug, c.name_uz, c.name_ru, c.name_en
	FROM products p
	LEFT JOIN categories c ON c.id = p.category_id`

// minPriceSQL orders by the cheapest in-stock volume.
const minPriceSQL = `(SELECT min(v.price) FROM product_variants v WHERE v.product_id = p.id AND v.in_stock)`

type scanner interface {
	Scan(dest ...any) error
}

func scanProduct(row scanner) (Product, error) {
	var p Product
	var cID *int64
	var cSlug, cUz, cRu, cEn *string
	err := row.Scan(&p.ID, &p.CategoryID, &p.NameUz, &p.NameRu, &p.NameEn,
		&p.DescriptionUz, &p.DescriptionRu, &p.DescriptionEn, &p.Brand, &p.Gender, &p.Concentration, &p.Tier,
		&p.NotesTop, &p.NotesHeart, &p.NotesBase,
		&p.ImageURL, &p.IsActive, &p.IsFeatured, &p.SortOrder,
		&p.CreatedAt, &p.UpdatedAt,
		&cID, &cSlug, &cUz, &cRu, &cEn)
	if err != nil {
		return p, err
	}
	if cID != nil {
		p.Category = &CategoryRef{ID: *cID, Slug: *cSlug, NameUz: *cUz, NameRu: *cRu, NameEn: *cEn}
	}
	p.Variants = []Variant{}
	return p, nil
}

// attachVariants loads the volumes of the given products in one query.
func (s *Store) attachVariants(ctx context.Context, products []Product) error {
	if len(products) == 0 {
		return nil
	}
	index := map[int64]*Product{}
	ids := make([]int64, len(products))
	for i := range products {
		ids[i] = products[i].ID
		index[products[i].ID] = &products[i]
	}
	rows, err := s.db.Query(ctx, `
		SELECT product_id, id, volume_ml, kind, price, in_stock, sort_order
		FROM product_variants WHERE product_id = ANY($1)
		ORDER BY sort_order, volume_ml, id`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var pid int64
		var v Variant
		if err := rows.Scan(&pid, &v.ID, &v.VolumeML, &v.Kind, &v.Price, &v.InStock, &v.SortOrder); err != nil {
			return err
		}
		p := index[pid]
		p.Variants = append(p.Variants, v)
		if v.InStock {
			p.InStock = true
			if p.MinPrice == nil || v.Price < *p.MinPrice {
				price := v.Price
				p.MinPrice = &price
			}
		}
	}
	return rows.Err()
}

func (s *Store) ListProducts(ctx context.Context, f ProductFilter) (Page[Product], error) {
	var where []string
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}

	if f.ActiveOnly {
		where = append(where, "p.is_active")
	}
	if f.CategoryID > 0 {
		where = append(where, "p.category_id = "+arg(f.CategoryID))
	}
	if f.CategorySlug != "" {
		where = append(where, "c.slug = "+arg(f.CategorySlug))
	}
	if ValidGender(f.Gender) {
		// Unisex scents belong in both the women's and the men's shelf.
		if f.Gender == "unisex" {
			where = append(where, "p.gender = 'unisex'")
		} else {
			where = append(where, "p.gender IN ("+arg(f.Gender)+", 'unisex')")
		}
	}
	if b := strings.TrimSpace(f.Brand); b != "" {
		where = append(where, "lower(p.brand) = lower("+arg(b)+")")
	}
	if ValidTier(f.Tier) {
		where = append(where, "p.tier = "+arg(f.Tier))
	}
	if f.Featured {
		where = append(where, "p.is_featured")
	}
	if len(f.IDs) > 0 {
		where = append(where, "p.id = ANY("+arg(f.IDs)+")")
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		like := arg("%" + escapeLike(q) + "%")
		where = append(where, fmt.Sprintf(
			"(p.name_uz ILIKE %[1]s OR p.name_ru ILIKE %[1]s OR p.name_en ILIKE %[1]s OR p.brand ILIKE %[1]s"+
				" OR p.notes_top ILIKE %[1]s OR p.notes_heart ILIKE %[1]s OR p.notes_base ILIKE %[1]s)", like))
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	orderSQL := " ORDER BY p.sort_order, p.id DESC"
	switch f.Sort {
	case "newest":
		orderSQL = " ORDER BY p.created_at DESC, p.id DESC"
	case "price_asc":
		orderSQL = " ORDER BY " + minPriceSQL + " ASC NULLS LAST, p.id DESC"
	case "price_desc":
		orderSQL = " ORDER BY " + minPriceSQL + " DESC NULLS LAST, p.id DESC"
	}

	page, limit := NormalizePage(f.Page, f.Limit, 100)
	result := Page[Product]{Items: []Product{}, Page: page, Limit: limit}

	if err := s.db.QueryRow(ctx,
		`SELECT count(*) FROM products p LEFT JOIN categories c ON c.id = p.category_id`+whereSQL, args...).
		Scan(&result.Total); err != nil {
		return result, err
	}

	query := productSelect + whereSQL + orderSQL +
		fmt.Sprintf(" LIMIT %d OFFSET %d", limit, (page-1)*limit)
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return result, err
	}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			rows.Close()
			return result, err
		}
		result.Items = append(result.Items, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return result, err
	}
	return result, s.attachVariants(ctx, result.Items)
}

func (s *Store) GetProduct(ctx context.Context, id int64) (Product, error) {
	p, err := scanProduct(s.db.QueryRow(ctx, productSelect+" WHERE p.id = $1", id))
	if err != nil {
		return p, notFound(err)
	}
	list := []Product{p}
	if err := s.attachVariants(ctx, list); err != nil {
		return p, err
	}
	return list[0], nil
}

type Brand struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ListBrands returns brands that have products on the storefront.
func (s *Store) ListBrands(ctx context.Context) ([]Brand, error) {
	rows, err := s.db.Query(ctx, `
		SELECT min(brand), count(*) FROM products
		WHERE is_active AND brand <> ''
		GROUP BY lower(brand)
		ORDER BY lower(min(brand))`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Brand{}
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.Name, &b.Count); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) CreateProduct(ctx context.Context, in ProductInput) (Product, error) {
	var id int64
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			INSERT INTO products (category_id, name_uz, name_ru, name_en, description_uz, description_ru,
				description_en, brand, gender, concentration, notes_top, notes_heart, notes_base,
				image_url, is_active, is_featured, sort_order, tier)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18) RETURNING id`,
			in.CategoryID, in.NameUz, in.NameRu, in.NameEn, in.DescriptionUz, in.DescriptionRu,
			in.DescriptionEn, in.Brand, in.Gender, in.Concentration, in.NotesTop, in.NotesHeart, in.NotesBase,
			in.ImageURL, in.IsActive, in.IsFeatured, in.SortOrder, in.Tier).Scan(&id); err != nil {
			return err
		}
		return saveVariants(ctx, tx, id, in.Variants)
	})
	if err != nil {
		return Product{}, err
	}
	return s.GetProduct(ctx, id)
}

func (s *Store) UpdateProduct(ctx context.Context, id int64, in ProductInput) (Product, error) {
	err := pgx.BeginFunc(ctx, s.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			UPDATE products SET category_id = $2, name_uz = $3, name_ru = $4, name_en = $5,
				description_uz = $6, description_ru = $7, description_en = $8, brand = $9, gender = $10,
				concentration = $11, notes_top = $12, notes_heart = $13, notes_base = $14,
				image_url = $15, is_active = $16, is_featured = $17, sort_order = $18, tier = $19, updated_at = now()
			WHERE id = $1`,
			id, in.CategoryID, in.NameUz, in.NameRu, in.NameEn, in.DescriptionUz, in.DescriptionRu,
			in.DescriptionEn, in.Brand, in.Gender, in.Concentration, in.NotesTop, in.NotesHeart, in.NotesBase,
			in.ImageURL, in.IsActive, in.IsFeatured, in.SortOrder, in.Tier)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrNotFound
		}
		return saveVariants(ctx, tx, id, in.Variants)
	})
	if err != nil {
		return Product{}, err
	}
	return s.GetProduct(ctx, id)
}

// saveVariants makes the product's volumes match the input: existing ids are
// updated (so past order lines keep their link), new ones inserted, missing ones removed.
func saveVariants(ctx context.Context, tx pgx.Tx, productID int64, variants []VariantInput) error {
	keep := []int64{}
	for i, v := range variants {
		if v.SortOrder == 0 {
			v.SortOrder = i
		}
		if v.ID > 0 {
			tag, err := tx.Exec(ctx, `
				UPDATE product_variants SET volume_ml = $3, kind = $4, price = $5, in_stock = $6, sort_order = $7
				WHERE id = $1 AND product_id = $2`,
				v.ID, productID, v.VolumeML, v.Kind, v.Price, v.InStock, v.SortOrder)
			if err != nil {
				return err
			}
			if tag.RowsAffected() > 0 {
				keep = append(keep, v.ID)
				continue
			}
		}
		var id int64
		if err := tx.QueryRow(ctx, `
			INSERT INTO product_variants (product_id, volume_ml, kind, price, in_stock, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			productID, v.VolumeML, v.Kind, v.Price, v.InStock, v.SortOrder).Scan(&id); err != nil {
			return err
		}
		keep = append(keep, id)
	}
	_, err := tx.Exec(ctx, `DELETE FROM product_variants WHERE product_id = $1 AND NOT (id = ANY($2))`, productID, keep)
	return err
}

func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
