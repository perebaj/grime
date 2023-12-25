// Package event provide the interface to handle events.
package event

import (
	"context"

	"gocloud.dev/pubsub"
)

type Event struct {
	Topic      string
	MaxHandler int
}

func NewEvent(topic string, maxHandler int) Event {
	return Event{
		Topic:      topic,
		MaxHandler: maxHandler,
	}
}

func (e Event) Send() error {
	ctx := context.Background()
	topic, err := pubsub.OpenTopic(ctx, e.Topic)
	if err != nil {
		return err
	}
	defer func() {
		_ = topic.Shutdown(ctx)
	}()

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte("Hello world!"),
	})

	if err != nil {
		return err
	}

	return nil
}

func (e Event) Receive(f func(msg *pubsub.Message) error, ctx context.Context) error {
	if e.MaxHandler == 0 || e.MaxHandler < 0 {
		e.MaxHandler = 10
	}

	subs, err := pubsub.OpenSubscription(ctx, e.Topic)
	if err != nil {
		return err
	}

	// defer func() {
	// 	_ = subs.Shutdown(ctx)
	// }()

	sem := make(chan struct{}, e.MaxHandler)
	for {
		msg, err := subs.Receive(ctx)
		if err != nil {
			return err
		}

		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			return ctx.Err()
		}

		go func() {
			defer func() { <-sem }() // release the semaphore
			defer msg.Ack()

			if err := f(msg); err != nil {
				return
			}
		}()
	}
}
