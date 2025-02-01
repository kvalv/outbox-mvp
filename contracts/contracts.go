package contracts

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kvalv/outbox-mvp/generated"
	"github.com/kvalv/outbox-mvp/publisher"
	"github.com/segmentio/ksuid"
)

type Employee struct {
	ID   string
	Name string
}

type Contract struct {
	ID            string
	Title         string
	Description   *string
	StartDate     time.Time
	Active        bool
	ResponsibleID *string
}

type ContractService struct {
	ps        publisher.Publisher
	db        *pgx.Conn
	dbQueries *generated.Queries
}

func NewService(ps publisher.Publisher, db *pgx.Conn, q *generated.Queries) *ContractService {
	return &ContractService{
		ps:        ps,
		db:        db,
		dbQueries: q,
	}
}

func (s *ContractService) Create(ctx context.Context, title string) (*Contract, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin query: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := s.dbQueries.WithTx(tx)
	cmd := generated.CreateContractParams{
		ID:    ksuid.New().String(),
		Title: title,
	}
	res, err := qtx.CreateContract(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract: %w", err)
	}

	data := []byte(fmt.Sprintf(`{"title": "%s"}`, title))
	evParams := generated.InsertEventParams{
		ID:         ksuid.New().String(),
		ContractID: cmd.ID,
		Type:       "contract-created",
		Data:       data,
	}
	if err := qtx.InsertEvent(ctx, evParams); err != nil {
		fmt.Printf("made it here %s\n", err)
		return nil, fmt.Errorf("failed to insert event: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit: %w", err)
	}
	// s.dbQueries.NotifyEvent(ctx, "event")

	return &Contract{
		ID:            res.ID,
		Title:         res.Title,
		Active:        res.Active,
		ResponsibleID: nil,
	}, nil
}

func (s *ContractService) Activate(ctx context.Context, contractID string, reason string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.dbQueries.WithTx(tx)
	contract, err := qtx.ActivateContract(ctx, contractID)
	if err != nil {
		return err
	}
	data := []byte(fmt.Sprintf(`{"reason": "%s"}`, reason))
	params := generated.InsertEventParams{
		ID:         ksuid.New().String(),
		ContractID: contract.ID,
		Type:       "contract-deactivated",
		Data:       data,
	}
	if err := qtx.InsertEvent(ctx, params); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// best effort notify that a new event is inserted
	s.dbQueries.NotifyEvent(ctx, "hello world")

	return tx.Commit(ctx)
}
