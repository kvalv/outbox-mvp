package relay

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/kvalv/outbox-mvp/ports"
	"github.com/kvalv/outbox-mvp/queries"
)

type EventRelay struct {
	conn *pgx.Conn
	ps   ports.Publisher
	q    *queries.Queries
	log  slog.Logger
}

func (r *EventRelay) Run(ctx context.Context, pollInterval time.Duration) error {
	t := time.Tick(pollInterval)
	notify := r.pgPoll(ctx)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t:
			r.log.Info("polling")
			r.poll(ctx)
		case <-notify:
			r.log.Info("notification")
			r.poll(ctx)
		}
	}
}

func (r *EventRelay) pgPoll(ctx context.Context) <-chan time.Time {
	ch := make(chan time.Time)
	go func() {
		defer close(ch)
		for {
			_, err := r.conn.WaitForNotification(ctx)
			if err != nil {
				return
			}
			select {
			case ch <- time.Now():
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch
}

func (r *EventRelay) poll(ctx context.Context) error {
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)
	events, err := qtx.GetUnsentEvents(ctx)
	if err != nil {
		return err
	}

	var count int
	for _, e := range events {
		r.log.Info("will publish event", "id", e.ID.String())
		if err := r.ps.Publish(ctx, e); err != nil {
			return err
		}
		if _, err := qtx.MarkEventAsProcessed(ctx, e.ID); err != nil {
			return err
		}
		count++
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if count > 0 {
		r.log.Info("published events", "count", count)
	}
	return nil
}
