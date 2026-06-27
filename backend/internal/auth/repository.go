package auth

import (
	"context"
	"database/sql"
)

type Repository interface {
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

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	const q = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, q, email).Scan(&exists)
	return exists, err
}
