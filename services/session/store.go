package session

import (
	"context"

	"github.com/PauloHInocencio/testers-admin-dashboard/db/database"
)

type Store struct {
	queries *database.Queries
}

type SessionStore interface {
	FindValidSession(ctx context.Context, tokenHash string) (string, error)
}

func NewStore(queries *database.Queries) *Store {
	return &Store{
		queries: queries,
	}
}

func (s *Store) FindValidSession(ctx context.Context, tokenHash string) (string, error) {
	return s.queries.FindValidSession(ctx, tokenHash)
}
