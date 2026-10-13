package store

import (
	"cmp"
	"context"
	"fmt"
	"strings"
	"time"
)

var OrderStatuses = []string{"new", "processing", "completed", "cancelled"}

func ValidOrderStatus(s string) bool {
	return oneOf(s, OrderStatuses)
}

type OrderItem struct {
	ID           int64  `json:"id"`
	ProductID    *int64 `json:"product_id"`
	VariantID    *int64 `json:"variant_id"`
	ProductName  string `json:"product_name"`
	ProductBrand string `json:"product_brand"`
	ProductImage string `json:"product_image"`
	VolumeML     int    `json:"volume_ml"`
	Kind         string `json:"kind"`
	Price        int64  `json:"price"`
	Quantity     int    `json:"quantity"`
	Subtotal     int64  `json:"subtotal"`
}

type Order struct {
	ID         int64       `json:"id"`
	FullName   string      `json:"full_name"`
	Phone      string      `json:"phone"`
	Comment    string      `json:"comment"`
	Status     string      `json:"status"`
	Total      int64       `json:"total"`
	AdminNote  string      `json:"admin_note"`
	Locale     string      `json:"locale"`
	Source     string      `json:"source"` // site | ai_chat
	ItemsCount int         `json:"items_count"`
	Items      []OrderItem `json:"items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItemInput struct {
	VariantID int64 `json:"variant_id"`
	Quantity  int   `json:"quantity"`
}

type OrderInput struct {
	FullName string           `json:"full_name"`
	Phone    string           `json:"phone"`
	Comment  string           `json:"comment"`
	Locale   string           `json:"locale"`
	Items    []OrderItemInput `json:"items"`
	// ChatSessionID links the order to the AI chat that led to it.
	ChatSessionID string `json:"chat_session_id"`
	Source        string `json:"-"` // "site" or "ai_chat"
}

type OrderFilter struct {
	Status string
	Query  string
	Page   int
	Limit  int
}

// CreateOrder prices every line from the database (never trusting client
// prices) and stores the order with a snapshot of each product and volume.
func (s *Store) CreateOrder(ctx context.Context, in OrderInput) (Order, error) {
	quantities := map[int64]int{}
	var ids []int64
	for _, it := range in.Items {
		if _, seen := quantities[it.VariantID]; !seen {
			ids = append(ids, it.VariantID)
		}
		quantities[it.VariantID] += it.Quantity
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT v.id, v.product_id, v.volume_ml, v.kind, v.price, p.name_uz, p.brand, p.image_url
		FROM product_variants v
		JOIN products p ON p.id = v.product_id
		WHERE v.id = ANY($1) AND v.in_stock AND p.is_active`, ids)
	if err != nil {
		return Order{}, err
	}
	found := map[int64]OrderItem{}
	for rows.Next() {
		var it OrderItem
		var vid, pid int64
		if err := rows.Scan(&vid, &pid, &it.VolumeML, &it.Kind, &it.Price, &it.ProductName, &it.ProductBrand, &it.ProductImage); err != nil {
			rows.Close()
			return Order{}, err
		}
		it.VariantID, it.ProductID = &vid, &pid
		found[vid] = it
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return Order{}, err
	}

	var items []OrderItem
	var total int64
	for _, id := range ids {
		it, ok := found[id]
		if !ok {
			return Order{}, fmt.Errorf("%w: variant %d is not available", ErrValidation, id)
		}
		it.Quantity = quantities[id]
		it.Subtotal = it.Price * int64(it.Quantity)
		total += it.Subtotal
		items = append(items, it)
	}

	var orderID int64
	if err := tx.QueryRow(ctx, `
		INSERT INTO orders (full_name, phone, comment, total, locale, source)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		in.FullName, in.Phone, in.Comment, total, in.Locale, cmp.Or(in.Source, "site")).Scan(&orderID); err != nil {
		return Order{}, err
	}
	if in.ChatSessionID != "" {
		if _, err := tx.Exec(ctx, `
			UPDATE chat_sessions SET order_id = $2, pending_order = NULL, updated_at = now()
			WHERE id = $1 AND order_id IS NULL`, in.ChatSessionID, orderID); err != nil {
			return Order{}, err
		}
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx, `
			INSERT INTO order_items (order_id, product_id, variant_id, product_name, product_brand, product_image,
				volume_ml, kind, price, quantity, subtotal)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			orderID, it.ProductID, it.VariantID, it.ProductName, it.ProductBrand, it.ProductImage,
			it.VolumeML, it.Kind, it.Price, it.Quantity, it.Subtotal); err != nil {
			return Order{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Order{}, err
	}
	return s.GetOrder(ctx, orderID)
}

const orderColumns = `o.id, o.full_name, o.phone, o.comment, o.status, o.total, o.admin_note, o.locale, o.source,
	o.created_at, o.updated_at,
	(SELECT coalesce(sum(quantity), 0) FROM order_items i WHERE i.order_id = o.id)`

func scanOrder(row scanner) (Order, error) {
	var o Order
	err := row.Scan(&o.ID, &o.FullName, &o.Phone, &o.Comment, &o.Status, &o.Total, &o.AdminNote,
		&o.Locale, &o.Source, &o.CreatedAt, &o.UpdatedAt, &o.ItemsCount)
	return o, err
}

func (s *Store) ListOrders(ctx context.Context, f OrderFilter) (Page[Order], error) {
	var where []string
	var args []any
	if f.Status != "" {
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("o.status = $%d", len(args)))
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(o.full_name ILIKE $%d OR o.phone ILIKE $%d OR o.id::text = $%d)", n, n, n+1))
		args = append(args, strings.TrimPrefix(q, "#"))
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	page, limit := NormalizePage(f.Page, f.Limit, 100)
	result := Page[Order]{Items: []Order{}, Page: page, Limit: limit}
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM orders o`+whereSQL, args...).Scan(&result.Total); err != nil {
		return result, err
	}

	rows, err := s.db.Query(ctx, `SELECT `+orderColumns+` FROM orders o`+whereSQL+
		fmt.Sprintf(" ORDER BY o.created_at DESC, o.id DESC LIMIT %d OFFSET %d", limit, (page-1)*limit), args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, o)
	}
	return result, rows.Err()
}

func (s *Store) GetOrder(ctx context.Context, id int64) (Order, error) {
	o, err := scanOrder(s.db.QueryRow(ctx, `SELECT `+orderColumns+` FROM orders o WHERE o.id = $1`, id))
	if err != nil {
		return o, notFound(err)
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, product_id, variant_id, product_name, product_brand, product_image, volume_ml, kind,
			price, quantity, subtotal
		FROM order_items WHERE order_id = $1 ORDER BY id`, id)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	o.Items = []OrderItem{}
	for rows.Next() {
		var it OrderItem
		if err := rows.Scan(&it.ID, &it.ProductID, &it.VariantID, &it.ProductName, &it.ProductBrand,
			&it.ProductImage, &it.VolumeML, &it.Kind, &it.Price, &it.Quantity, &it.Subtotal); err != nil {
			return o, err
		}
		o.Items = append(o.Items, it)
	}
	return o, rows.Err()
}

func (s *Store) UpdateOrder(ctx context.Context, id int64, status, adminNote string) (Order, error) {
	tag, err := s.db.Exec(ctx, `UPDATE orders SET status = $2, admin_note = $3, updated_at = now() WHERE id = $1`,
		id, status, adminNote)
	if err != nil {
		return Order{}, err
	}
	if tag.RowsAffected() == 0 {
		return Order{}, ErrNotFound
	}
	return s.GetOrder(ctx, id)
}

func (s *Store) DeleteOrder(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM orders WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
