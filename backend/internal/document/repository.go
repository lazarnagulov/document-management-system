package document

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"bitbucket.org/HyperEther/document-management-system/internal/platform/pagination"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id uuid.UUID) (*Document, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, page pagination.Page) ([]*Document, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, doc *Document) error {
	const q = `
		INSERT INTO documents (id, user_id, name, content_type, size, storage_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, q,
		doc.ID, doc.UserID, doc.Name, doc.ContentType,
		doc.Size, doc.StorageKey, doc.CreatedAt, doc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create document: %w", err)
	}
	return nil
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	const q = `
		SELECT id, user_id, name, content_type, size, storage_key, created_at, updated_at
		FROM documents WHERE id = $1`

	d := &Document{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&d.ID, &d.UserID, &d.Name, &d.ContentType,
		&d.Size, &d.StorageKey, &d.CreatedAt, &d.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get document: %w", err)
	}
	return d, nil
}

func (r *repository) Update(ctx context.Context, doc *Document) error {
	const q = `UPDATE documents SET name = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, q, doc.Name, doc.UpdatedAt, doc.ID)
	if err != nil {
		return fmt.Errorf("update document: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update document rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete document rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *repository) ListByUserID(ctx context.Context, userID uuid.UUID, page pagination.Page) ([]*Document, error) {
	const q = `
		SELECT id, user_id, name, content_type, size, storage_key, created_at, updated_at
		FROM documents WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, q, userID, page.Size, page.Offset())
	if err != nil {
		return nil, fmt.Errorf("list documents: %w", err)
	}
	defer rows.Close()

	var docs []*Document
	for rows.Next() {
		d := &Document{}
		if err := rows.Scan(
			&d.ID, &d.UserID, &d.Name, &d.ContentType,
			&d.Size, &d.StorageKey, &d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan document: %w", err)
		}
		docs = append(docs, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list documents rows: %w", err)
	}
	return docs, nil
}

func (r *repository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	const q = `SELECT COUNT(*) FROM documents WHERE user_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, q, userID).Scan(&total); err != nil {
		return 0, fmt.Errorf("count documents: %w", err)
	}
	return total, nil
}
