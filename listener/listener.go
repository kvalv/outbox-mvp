package listener

import (
	"context"
	"fmt"
	"log/slog"
)

type listener struct {
	log *slog.Logger
}

func New(log *slog.Logger) *listener {
	return &listener{log: log}
}

func (l *listener) Listen(ctx context.Context, ch <-chan []byte) {
	l.log.Info("listening")
	for {
		select {
		case <-ctx.Done():
			l.log.Info("hanging up")
			return
		case msg := <-ch:
			data := string(msg)
			l.log.Info(fmt.Sprintf("received message: %s", data))
		}
	}
}
