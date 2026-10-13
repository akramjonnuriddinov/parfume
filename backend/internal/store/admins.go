package store

import (
	"context"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
}

// EnsureAdmin creates the initial admin account when no admin exists yet.
func (s *Store) EnsureAdmin(ctx context.Context, username, password string) error {
	var count int
	if err := s.db.QueryRow(ctx, `SELECT count(*) FROM admins`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `INSERT INTO admins (username, password_hash) VALUES ($1, $2)`, username, string(hash))
	return err
}

func (s *Store) AdminByUsername(ctx context.Context, username string) (Admin, error) {
	var a Admin
	err := s.db.QueryRow(ctx, `SELECT id, username, password_hash FROM admins WHERE username = $1`, username).
		Scan(&a.ID, &a.Username, &a.PasswordHash)
	return a, notFound(err)
}

func (s *Store) AdminByID(ctx context.Context, id int64) (Admin, error) {
	var a Admin
	err := s.db.QueryRow(ctx, `SELECT id, username, password_hash FROM admins WHERE id = $1`, id).
		Scan(&a.ID, &a.Username, &a.PasswordHash)
	return a, notFound(err)
}

func (s *Store) UpdateAdminPassword(ctx context.Context, id int64, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `UPDATE admins SET password_hash = $2 WHERE id = $1`, id, string(hash))
	return err
}
