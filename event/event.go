// Package event provide the interface to handle events.
package event

import (
	"context"
	"fmt"

	"gocloud.dev/pubsub"
)

type Event struct {
	Topic string
}

func NewEvent(topic string) Event {
	return Event{
		Topic: topic,
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

func (e Event) Receive() error {
	ctx := context.Background()
	subs, err := pubsub.OpenSubscription(ctx, e.Topic)
	if err != nil {
		return err
	}

	defer func() {
		_ = subs.Shutdown(ctx)
	}()

	msg, err := subs.Receive(ctx)
	if err != nil {
		return err
	}
	fmt.Println(string(msg.Body))
	msg.Ack()

	return nil
}
