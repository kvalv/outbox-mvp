package publisher

import "context"

type channelPublisher struct {
	ch chan<- []byte
}

func (p *channelPublisher) Publish(ctx context.Context, msg []byte) error {
	p.ch <- msg
	return nil
}

func NewChannelPublisher(ch chan<- []byte) Publisher {
	return &channelPublisher{ch: ch}
}
