package auth

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type Repository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	Create(ctx context.Context, user *User) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	const q = `
        INSERT INTO users (id, email, password)
        VALUES ($1, $2, $3)`

	_, err := r.db.ExecContext(ctx, q,
		user.ID,
		user.Email,
		user.Password,
	)
	return err
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	const q = `
        SELECT id, email, password, created_at, updated_at
        FROM users WHERE email = $1`

	u := &User{}
	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	const q = `
        SELECT id, email, password, created_at, updated_at
        FROM users WHERE id = $1`

	u := &User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Email, &u.Password, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, q, email).Scan(&exists)
	return exists, err
}
