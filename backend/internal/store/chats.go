package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type ChatSession struct {
	ID            string          `json:"id"`
	Locale        string          `json:"locale"`
	FirstMessage  string          `json:"first_message"`
	MessagesCount int             `json:"messages_count"`
	LeadsCount    int             `json:"leads_count"`
	PendingOrder  json.RawMessage `json:"-"`
	OrderID       *int64          `json:"order_id"`
	OrderStatus   *string         `json:"order_status"`
	OrderTotal    *int64          `json:"order_total"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type ChatMessage struct {
	ID        int64           `json:"id"`
	Role      string          `json:"role"`
	Content   string          `json:"content"`
	Meta      json.RawMessage `json:"meta"`
	CreatedAt time.Time       `json:"created_at"`
}

type ChatSessionDetail struct {
	ChatSession
	Messages []ChatMessage `json:"messages"`
	Leads    []Message     `json:"leads"`
}

type ChatFilter struct {
	Outcome string // "" | order | lead
	Query   string
	Page    int
	Limit   int
}

type MissingProduct struct {
	Name       string    `json:"name"`
	Count      int       `json:"count"`
	LastAsked  time.Time `json:"last_asked"`
	Details    []string  `json:"details"`
	SessionIDs []string  `json:"session_ids"`
}

const chatSessionSelect = `
	SELECT s.id, s.locale, s.first_message, s.messages_count, s.leads_count, s.pending_order,
		s.order_id, o.status, o.total, s.created_at, s.updated_at
	FROM chat_sessions s
	LEFT JOIN orders o ON o.id = s.order_id`

func scanChatSession(row scanner) (ChatSession, error) {
	var c ChatSession
	var pending []byte
	err := row.Scan(&c.ID, &c.Locale, &c.FirstMessage, &c.MessagesCount, &c.LeadsCount, &pending,
		&c.OrderID, &c.OrderStatus, &c.OrderTotal, &c.CreatedAt, &c.UpdatedAt)
	c.PendingOrder = pending
	return c, err
}

// EnsureChatSession returns the session with the given id, creating it when
// it doesn't exist yet.
func (s *Store) EnsureChatSession(ctx context.Context, id, locale string) (ChatSession, error) {
	if _, err := s.db.Exec(ctx, `
		INSERT INTO chat_sessions (id, locale) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, id, locale); err != nil {
		return ChatSession{}, err
	}
	return s.GetChatSession(ctx, id)
}

func (s *Store) GetChatSession(ctx context.Context, id string) (ChatSession, error) {
	c, err := scanChatSession(s.db.QueryRow(ctx, chatSessionSelect+` WHERE s.id = $1`, id))
	return c, notFound(err)
}

// RecentChatMessages returns the latest messages of a session in chronological order.
func (s *Store) RecentChatMessages(ctx context.Context, sessionID string, limit int) ([]ChatMessage, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, role, content, meta, created_at FROM (
			SELECT * FROM chat_messages WHERE session_id = $1 ORDER BY id DESC LIMIT $2
		) m ORDER BY id`, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ChatMessage{}
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.Role, &m.Content, &m.Meta, &m.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// AppendChatMessage stores one message and bumps the session counters.
func (s *Store) AppendChatMessage(ctx context.Context, sessionID, role, content string, meta any) error {
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if meta == nil {
		metaJSON = []byte("{}")
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO chat_messages (session_id, role, content, meta) VALUES ($1, $2, $3, $4)`,
		sessionID, role, content, metaJSON); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE chat_sessions SET messages_count = messages_count + 1, updated_at = now(),
			first_message = CASE WHEN first_message = '' AND $2 = 'user' THEN left($3, 300) ELSE first_message END
		WHERE id = $1`, sessionID, role, content); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// SetPendingOrder stores (or clears, with nil) the order awaiting confirmation.
func (s *Store) SetPendingOrder(ctx context.Context, sessionID string, pending any) error {
	var raw []byte
	if pending != nil {
		var err error
		if raw, err = json.Marshal(pending); err != nil {
			return err
		}
	}
	_, err := s.db.Exec(ctx, `UPDATE chat_sessions SET pending_order = $2 WHERE id = $1`, sessionID, raw)
	return err
}

func (s *Store) IncrementChatLeads(ctx context.Context, sessionID string) error {
	_, err := s.db.Exec(ctx, `UPDATE chat_sessions SET leads_count = leads_count + 1 WHERE id = $1`, sessionID)
	return err
}

func (s *Store) AddMissingProduct(ctx context.Context, sessionID, name, details string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO chat_missing_products (session_id, name, details) VALUES (NULLIF($1, ''), $2, $3)`,
		sessionID, name, details)
	return err
}

// --- Admin ---

func (s *Store) ListChatSessions(ctx context.Context, f ChatFilter) (Page[ChatSession], error) {
	where := []string{"s.messages_count > 0"}
	var args []any
	switch f.Outcome {
	case "order":
		where = append(where, "s.order_id IS NOT NULL")
	case "lead":
		where = append(where, "s.leads_count > 0")
	}
	if q := strings.TrimSpace(f.Query); q != "" {
		args = append(args, "%"+escapeLike(q)+"%")
		where = append(where, fmt.Sprintf(
			"EXISTS (SELECT 1 FROM chat_messages m WHERE m.session_id = s.id AND m.content ILIKE $%d)", len(args)))
	}
	whereSQL := " WHERE " + strings.Join(where, " AND ")

	page, limit := NormalizePage(f.Page, f.Limit, 100)
	result := Page[ChatSession]{Items: []ChatSession{}, Page: page, Limit: limit}
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM chat_sessions s`+whereSQL, args...).
		Scan(&result.Total); err != nil {
		return result, err
	}
	rows, err := s.db.Query(ctx, chatSessionSelect+whereSQL+
		fmt.Sprintf(" ORDER BY s.updated_at DESC LIMIT %d OFFSET %d", limit, (page-1)*limit), args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		c, err := scanChatSession(rows)
		if err != nil {
			return result, err
		}
		result.Items = append(result.Items, c)
	}
	return result, rows.Err()
}

func (s *Store) GetChatSessionDetail(ctx context.Context, id string) (ChatSessionDetail, error) {
	c, err := s.GetChatSession(ctx, id)
	if err != nil {
		return ChatSessionDetail{}, err
	}
	d := ChatSessionDetail{ChatSession: c}
	if d.Messages, err = s.RecentChatMessages(ctx, id, 1000); err != nil {
		return d, err
	}

	rows, err := s.db.Query(ctx, `SELECT `+messageColumns+` FROM messages WHERE chat_session_id = $1 ORDER BY id`, id)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	d.Leads = []Message{}
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return d, err
		}
		d.Leads = append(d.Leads, m)
	}
	return d, rows.Err()
}

func (s *Store) DeleteChatSession(ctx context.Context, id string) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM chat_sessions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMissingProducts groups unmet requests by name, most requested first.
func (s *Store) ListMissingProducts(ctx context.Context, days int) ([]MissingProduct, error) {
	rows, err := s.db.Query(ctx, `
		SELECT min(name), count(*), max(created_at),
			(array_remove(array_agg(NULLIF(details, '') ORDER BY created_at DESC), NULL))[1:5],
			(array_remove(array_agg(DISTINCT session_id), NULL))[1:20]
		FROM chat_missing_products
		WHERE $1 <= 0 OR created_at >= now() - make_interval(days => $1)
		GROUP BY lower(name)
		ORDER BY count(*) DESC, max(created_at) DESC
		LIMIT 200`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MissingProduct{}
	for rows.Next() {
		var m MissingProduct
		if err := rows.Scan(&m.Name, &m.Count, &m.LastAsked, &m.Details, &m.SessionIDs); err != nil {
			return nil, err
		}
		if m.Details == nil {
			m.Details = []string{}
		}
		if m.SessionIDs == nil {
			m.SessionIDs = []string{}
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

func (s *Store) DeleteMissingProduct(ctx context.Context, name string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM chat_missing_products WHERE lower(name) = lower($1)`, name)
	return err
}
