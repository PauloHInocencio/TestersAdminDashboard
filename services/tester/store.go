package tester

import (
	"context"

	"github.com/PauloHInocencio/testers-admin-dashboard/db/database"
	"github.com/PauloHInocencio/testers-admin-dashboard/models"
	"github.com/google/uuid"
)

type Store struct {
	queries *database.Queries
}

type TestersStore interface {
	CreateSignup(ctx context.Context, email string, name string, platform string) error
	ListAll(ctx context.Context) ([]models.TesterSignup, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.TesterSignup, error)
	Approve(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	MarkAsInvited(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

func NewStore(queries *database.Queries) *Store {
	return &Store{
		queries: queries,
	}
}

func (s *Store) CreateSignup(ctx context.Context, email string, name string, platform string) error {
	return s.queries.CreateSignup(ctx, database.CreateSignupParams{
		Email:    email,
		Name:     name,
		Platform: platform,
	})
}

func (s *Store) ListAll(ctx context.Context) ([]models.TesterSignup, error) {
	rows, err := s.queries.ListTesters(ctx)
	if err != nil {
		return nil, err
	}

	testers := make([]models.TesterSignup, 0, len(rows))
	for _, row := range rows {
		testers = append(testers, toModel(row))
	}

	return testers, nil
}

func (s *Store) FindByID(ctx context.Context, id uuid.UUID) (models.TesterSignup, error) {
	row, err := s.queries.FindTesterByID(ctx, id)
	return toModel(row), err
}

func (s *Store) Approve(ctx context.Context, id uuid.UUID) error {
	return s.queries.ApproveTester(ctx, id)
}

func (s *Store) Reject(ctx context.Context, id uuid.UUID) error {
	return s.queries.RejectTester(ctx, id)
}

func (s *Store) MarkAsInvited(ctx context.Context, id uuid.UUID) error {
	return s.queries.MarkTesterInvited(ctx, id)
}

func (s *Store) Delete(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteTester(ctx, id)
}

func toModel(row database.TesterSignup) models.TesterSignup {
	return models.TesterSignup{
		ID:         row.ID.String(),
		Email:      row.Email,
		Name:       row.Name,
		Platform:   row.Platform,
		Status:     row.Status,
		CreatedAt:  row.CreatedAt,
		ApprovedAt: &row.ApprovedAt.Time,
		RejectedAt: &row.RejectedAt.Time,
		InvitedAt:  &row.InvitedAt.Time,
	}
}
