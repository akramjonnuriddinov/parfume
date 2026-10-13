package store

import (
	"cmp"
	"context"
	"fmt"
	"time"
)

type Message struct {
	ID          int64     `json:"id"`
	FullName    string    `json:"full_name"`
	Phone       string    `json:"phone"`
	Message     string    `json:"message"`
	ProductID   *int64    `json:"product_id"`
	ProductName string    `json:"product_name"`
	IsRead      bool      `json:"is_read"`
	Source      string    `json:"source"` // site | ai_callback | ai_quote
	ChatID      *string   `json:"chat_session_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type MessageInput struct {
	FullName  string `json:"full_name"`
	Phone     string `json:"phone"`
	Message   string `json:"message"`
	ProductID *int64 `json:"product_id"`
	Source    string `json:"-"`
	ChatID    string `json:"-"`
}

const messageColumns = `id, full_name, phone, message, product_id, product_name, is_read, source, chat_session_id, created_at`

func scanMessage(row scanner) (Message, error) {
	var m Message
	err := row.Scan(&m.ID, &m.FullName, &m.Phone, &m.Message, &m.ProductID, &m.ProductName, &m.IsRead,
		&m.Source, &m.ChatID, &m.CreatedAt)
	return m, err
}

func (s *Store) CreateMessage(ctx context.Context, in MessageInput) (Message, error) {
	productName := ""
	if in.ProductID != nil {
		if err := s.db.QueryRow(ctx, `SELECT trim(brand || ' ' || name_uz) FROM products WHERE id = $1`, *in.ProductID).
			Scan(&productName); err != nil {
			in.ProductID = nil
		}
	}
	return scanMessage(s.db.QueryRow(ctx, `
		INSERT INTO messages (full_name, phone, message, product_id, product_name, source, chat_session_id)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
		RETURNING `+messageColumns,
		in.FullName, in.Phone, in.Message, in.ProductID, productName, cmp.Or(in.Source, "site"), in.ChatID))
}

func (s *Store) ListMessages(ctx context.Context, unreadOnly bool, page, limit int) (Page[Message], error) {
	page, limit = NormalizePage(page, limit, 100)
	result := Page[Message]{Items: []Message{}, Page: page, Limit: limit}

	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM messages WHERE NOT $1 OR NOT is_read`, unreadOnly).
		Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := s.db.Query(ctx, `
		SELECT `+messageColumns+`
		FROM messages WHERE NOT $1 OR NOT is_read
		ORDER BY created_at DESC, id DESC`+fmt.Sprintf(" LIMIT %d OFFSET %d", limit, (page-1)*limit), unreadOnly)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, m)
	}
	return result, rows.Err()
}

func (s *Store) SetMessageRead(ctx context.Context, id int64, read bool) error {
	tag, err := s.db.Exec(ctx, `UPDATE messages SET is_read = $2 WHERE id = $1`, id, read)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) DeleteMessage(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM messages WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
