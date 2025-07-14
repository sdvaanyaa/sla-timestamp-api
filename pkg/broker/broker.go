package broker

import "context"

type Broker interface {
	Publish(ctx context.Context, msg []byte) error
	Close() error
}
