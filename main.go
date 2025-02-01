package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kvalv/outbox-mvp/contracts"
	"github.com/kvalv/outbox-mvp/generated"
	"github.com/kvalv/outbox-mvp/listener"
	"github.com/kvalv/outbox-mvp/publisher"
	"github.com/kvalv/outbox-mvp/relay"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	handle := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	log := slog.New(handle)

	pgDsn := "postgres://postgres:postgres@localhost:5432/postgres"

	pool, err := pgxpool.New(context.Background(), pgDsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to connect to database:", err)
		os.Exit(1)
	}

	ch := make(chan []byte)
	ps := publisher.NewChannelPublisher(ch)
	q := generated.New(pool)
	eventRelay := relay.New("ev", pool, ps, q, log.With("comp", "EventRelay"))
	eventListener := listener.New(log.With("comp", "eventListener"))

	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Error("failed to acquire connection", "err", err.Error())
		return
	}
	defer conn.Release()
	service := contracts.NewService(ps, conn.Hijack(), q)

	go eventRelay.Run(ctx)
	go eventListener.Listen(ctx, ch)

	go func() {
		// execute some work after everything has started
		time.Sleep(300 * time.Millisecond)
		contract, err := service.Create(ctx, "hello")
		if err != nil {
			log.Error("failed to create contract", "err", err.Error())
			return
		}
		time.Sleep(500 * time.Millisecond)
		if err := service.Activate(ctx, contract.ID, "company went bankrupt"); err != nil {
			log.Error("failed to activate contract",
				"err", err.Error(),
				"id", contract.ID,
			)
			return
		}
	}()

	<-ctx.Done()
}
