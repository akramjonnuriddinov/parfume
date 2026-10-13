package store

import "context"

// SettingKeys is the whitelist of editable site settings.
var SettingKeys = []string{
	"phone", "phone_secondary", "email", "telegram_url", "instagram_url",
	"address_uz", "address_ru", "address_en",
	"work_hours_uz", "work_hours_ru", "work_hours_en",
}

func (s *Store) GetSettings(ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	for _, k := range SettingKeys {
		out[k] = ""
	}
	rows, err := s.db.Query(ctx, `SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		if _, ok := out[k]; ok {
			out[k] = v
		}
	}
	return out, rows.Err()
}

func (s *Store) SaveSettings(ctx context.Context, values map[string]string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, k := range SettingKeys {
		v, ok := values[k]
		if !ok {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO settings (key, value) VALUES ($1, $2)
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value`, k, v); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type Stats struct {
	OrdersTotal     int            `json:"orders_total"`
	OrdersByStatus  map[string]int `json:"orders_by_status"`
	RevenueTotal    int64          `json:"revenue_total"`
	OrdersToday     int            `json:"orders_today"`
	ProductsTotal   int            `json:"products_total"`
	ProductsActive  int            `json:"products_active"`
	CategoriesTotal int            `json:"categories_total"`
	MessagesUnread  int            `json:"messages_unread"`
	RecentOrders    []Order        `json:"recent_orders"`
}

func (s *Store) GetStats(ctx context.Context) (Stats, error) {
	st := Stats{OrdersByStatus: map[string]int{}}
	for _, status := range OrderStatuses {
		st.OrdersByStatus[status] = 0
	}

	rows, err := s.db.Query(ctx, `SELECT status, count(*) FROM orders GROUP BY status`)
	if err != nil {
		return st, err
	}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			rows.Close()
			return st, err
		}
		st.OrdersByStatus[status] = n
		st.OrdersTotal += n
	}
	rows.Close()

	err = s.db.QueryRow(ctx, `
		SELECT
			(SELECT coalesce(sum(total), 0) FROM orders WHERE status = 'completed'),
			(SELECT count(*) FROM orders WHERE created_at >= date_trunc('day', now())),
			(SELECT count(*) FROM products),
			(SELECT count(*) FROM products WHERE is_active),
			(SELECT count(*) FROM categories),
			(SELECT count(*) FROM messages WHERE NOT is_read)`).
		Scan(&st.RevenueTotal, &st.OrdersToday, &st.ProductsTotal, &st.ProductsActive,
			&st.CategoriesTotal, &st.MessagesUnread)
	if err != nil {
		return st, err
	}

	recent, err := s.ListOrders(ctx, OrderFilter{Page: 1, Limit: 6})
	if err != nil {
		return st, err
	}
	st.RecentOrders = recent.Items
	return st, nil
}
