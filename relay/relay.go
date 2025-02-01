// relay is responsible for listening to events stored in the events table
// and publishing them to a message broker when they are ready to be sent.
// Once they are sent, the events are marked as processed in the database.
//
// In a sense, they relay events from the database to the message broker.

package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kvalv/outbox-mvp/generated"
	"github.com/kvalv/outbox-mvp/publisher"
)

type EventRelay struct {
	channel string
	conn    *pgxpool.Pool
	ps      publisher.Publisher
	q       *generated.Queries
	log     *slog.Logger
}

func New(
	channel string,
	conn *pgxpool.Pool,
	ps publisher.Publisher,
	q *generated.Queries,
	log *slog.Logger,
) *EventRelay {
	return &EventRelay{
		channel: channel,
		conn:    conn,
		ps:      ps,
		q:       q,
		log:     log,
	}
}

func (r *EventRelay) Run(ctx context.Context) error {
	notify := r.listen(ctx)
	r.log.Info("listening")
	for {
		select {
		case <-notify:
			r.log.Info("notification received")
			if err := r.publishUnsentEvents(ctx); err != nil {
				r.log.Error("failed to poll", "err", err.Error())
			}
		case <-ctx.Done():
			r.log.Info("Done - exiting")
			return nil
		}
	}
}

// listen waits for notifications from the postgres database. When they come,
// an event with the timestamp of when the notification is received is sent to
// the returned channel.
func (r *EventRelay) listen(ctx context.Context) <-chan time.Time {
	ch := make(chan time.Time)
	done := ctx.Done()

	conn, err := r.conn.Acquire(ctx)
	if err != nil {
		panic(fmt.Sprintf("failed to acquire connection: %v", err))
	}

	if _, err := conn.Exec(ctx, fmt.Sprintf("LISTEN %s", r.channel)); err != nil {
		r.log.Error("failed to listen", "err", err.Error())
		return nil
	}

	pgxConn := conn.Hijack()

	go func() {
		defer close(ch)
		for {
			_, err := pgxConn.WaitForNotification(ctx)
			if err != nil {
				return
			}
			select {
			case <-done:
				return
			case ch <- time.Now():
			}
		}
	}()
	return ch
}

// an event that gets published as a JSON message to the message broker
type event struct {
	ID         string    `json:"id"`
	ContractID string    `json:"contract_id"`
	Type       string    `json:"type"`
	TStamp     time.Time `json:"tstamp"`
	Data       []byte    `json:"data"`
}

// publishUnsentEvents fetches any events that is not yet published in the database, and
// publishes them. On success, the events are marked as published, and they will not be
// re-published in the future.
func (r *EventRelay) publishUnsentEvents(ctx context.Context) error {
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
		r.log.Info("will publish event", "id", e.ID)

		bytes, _ := json.Marshal(&event{
			ID:         e.ID,
			ContractID: e.ContractID,
			Type:       e.Type,
			TStamp:     e.Ts.Time.UTC(),
			Data:       e.Data,
		})

		if err := r.ps.Publish(ctx, bytes); err != nil {
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
