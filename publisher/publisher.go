package publisher

import "context"

type Publisher interface {
	Publish(ctx context.Context, msg []byte) error
}
