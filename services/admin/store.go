package admin

import (
	"context"
	"database/sql"
	"time"

	"github.com/PauloHInocencio/testers-admin-dashboard/db/database"
	"github.com/google/uuid"
)

type Store struct {
	queries *database.Queries
	db      *sql.DB
}

type AdminAuthParams struct {
	id        uuid.UUID
	email     string
	tokenHash string
	expiresAt time.Time
}

type AdminStore interface {
	IsAdminWhitelisted(ctx context.Context, email string) (bool, error)
	CreateMagicLink(ctx context.Context, params AdminAuthParams) error
	ConsumeMagicLink(ctx context.Context, tokenHash string) (string, error)
	CreateAdminSession(ctx context.Context, params AdminAuthParams) error
	FindValidSession(ctx context.Context, tokenHash string) (string, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}

func NewStore(queries *database.Queries, db *sql.DB) *Store {
	return &Store{
		queries: queries,
		db:      db,
	}
}

func (s *Store) IsAdminWhitelisted(ctx context.Context, email string) (bool, error) {
	return s.queries.IsAdminWhitelisted(ctx, email)
}

func (s *Store) CreateMagicLink(ctx context.Context, params AdminAuthParams) error {
	return s.queries.CreateMagicLink(ctx, database.CreateMagicLinkParams{
		ID:        params.id,
		Email:     params.email,
		TokenHash: params.tokenHash,
		ExpiresAt: params.expiresAt,
	})

}

func (s *Store) ConsumeMagicLink(ctx context.Context, tokenHash string) (string, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	qtx := s.queries.WithTx(tx)

	email, err := qtx.FindValidMagicLinkForUpdate(ctx, tokenHash)
	if err != nil {
		return "", err
	}

	if err = qtx.MarkMagicLinkUsed(ctx, tokenHash); err != nil {
		return "", err
	}

	if err = tx.Commit(); err != nil {
		return "", err
	}

	return email, nil

}

func (s *Store) CreateAdminSession(ctx context.Context, params AdminAuthParams) error {
	return s.queries.CreateAdminSession(ctx, database.CreateAdminSessionParams{
		ID:        params.id,
		Email:     params.email,
		TokenHash: params.tokenHash,
		ExpiresAt: params.expiresAt,
	})
}

func (s *Store) FindValidSession(ctx context.Context, tokenHash string) (string, error) {
	return s.queries.FindValidSession(ctx, tokenHash)
}
func (s *Store) DeleteSession(ctx context.Context, tokenHash string) error {
	return s.queries.DeleteSession(ctx, tokenHash)
}
