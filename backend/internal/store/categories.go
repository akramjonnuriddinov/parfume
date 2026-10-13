package store

import (
	"context"
	"time"
)

type Category struct {
	ID            int64     `json:"id"`
	Slug          string    `json:"slug"`
	NameUz        string    `json:"name_uz"`
	NameRu        string    `json:"name_ru"`
	NameEn        string    `json:"name_en"`
	DescriptionUz string    `json:"description_uz"`
	DescriptionRu string    `json:"description_ru"`
	DescriptionEn string    `json:"description_en"`
	SortOrder     int       `json:"sort_order"`
	ProductCount  int       `json:"product_count"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CategoryInput struct {
	Slug          string `json:"slug"`
	NameUz        string `json:"name_uz"`
	NameRu        string `json:"name_ru"`
	NameEn        string `json:"name_en"`
	DescriptionUz string `json:"description_uz"`
	DescriptionRu string `json:"description_ru"`
	DescriptionEn string `json:"description_en"`
	SortOrder     int    `json:"sort_order"`
}

const categoryColumns = `c.id, c.slug, c.name_uz, c.name_ru, c.name_en,
	c.description_uz, c.description_ru, c.description_en, c.sort_order,
	c.created_at, c.updated_at`

// ListCategories returns all categories; when activeOnly is set the product
// count only includes products visible on the storefront.
func (s *Store) ListCategories(ctx context.Context, activeOnly bool) ([]Category, error) {
	rows, err := s.db.Query(ctx, `
		SELECT `+categoryColumns+`,
			(SELECT count(*) FROM products p WHERE p.category_id = c.id AND (NOT $1 OR p.is_active))
		FROM categories c
		ORDER BY c.sort_order, c.id`, activeOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.NameUz, &c.NameRu, &c.NameEn,
			&c.DescriptionUz, &c.DescriptionRu, &c.DescriptionEn, &c.SortOrder,
			&c.CreatedAt, &c.UpdatedAt, &c.ProductCount); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

func (s *Store) GetCategory(ctx context.Context, id int64) (Category, error) {
	var c Category
	err := s.db.QueryRow(ctx, `SELECT `+categoryColumns+` FROM categories c WHERE c.id = $1`, id).
		Scan(&c.ID, &c.Slug, &c.NameUz, &c.NameRu, &c.NameEn,
			&c.DescriptionUz, &c.DescriptionRu, &c.DescriptionEn, &c.SortOrder,
			&c.CreatedAt, &c.UpdatedAt)
	return c, notFound(err)
}

func (s *Store) CreateCategory(ctx context.Context, in CategoryInput) (Category, error) {
	var id int64
	err := s.db.QueryRow(ctx, `
		INSERT INTO categories (slug, name_uz, name_ru, name_en, description_uz, description_ru, description_en, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		in.Slug, in.NameUz, in.NameRu, in.NameEn, in.DescriptionUz, in.DescriptionRu, in.DescriptionEn, in.SortOrder).
		Scan(&id)
	if err != nil {
		return Category{}, err
	}
	return s.GetCategory(ctx, id)
}

func (s *Store) UpdateCategory(ctx context.Context, id int64, in CategoryInput) (Category, error) {
	tag, err := s.db.Exec(ctx, `
		UPDATE categories SET slug = $2, name_uz = $3, name_ru = $4, name_en = $5,
			description_uz = $6, description_ru = $7, description_en = $8, sort_order = $9, updated_at = now()
		WHERE id = $1`,
		id, in.Slug, in.NameUz, in.NameRu, in.NameEn, in.DescriptionUz, in.DescriptionRu, in.DescriptionEn, in.SortOrder)
	if err != nil {
		return Category{}, err
	}
	if tag.RowsAffected() == 0 {
		return Category{}, ErrNotFound
	}
	return s.GetCategory(ctx, id)
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
