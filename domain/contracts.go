package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/kvalv/outbox-mvp/ports"
	"github.com/kvalv/outbox-mvp/queries"
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
	ps ports.Publisher
	db *pgx.Conn
	q  *queries.Queries
}

func (s *ContractService) DeactivateContract(ctx context.Context, contractID string, reason string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := s.q.WithTx(tx)
	contract, err := qtx.DeactivateContract(ctx, uuid.New())
	if err != nil {
		return err
	}
	data := []byte(fmt.Sprintf(`{"reason": "%s"}`, reason))
	if _, err := qtx.CreateContractEvent(ctx, queries.CreateContractEventParams{
		ContractID: contract.ID.String(),
		Type:       "contract-deactivated",
		Data:       data,
	}); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	// best effort notify that a new event is inserted
	s.q.NotifyEvent(ctx, "hello world")

	return tx.Commit(ctx)
}
