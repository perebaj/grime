// Package event is the entry point for the event package.
package event

import (
	"context"

	"gocloud.dev/pubsub"
)

// Publish publish an event to a topic
type Publish struct {
	topic *pubsub.Topic
}

// Subscription subscribe to a topic and receive messages
type Subscription struct {
	subs       *pubsub.Subscription
	MaxHandler int
}

// Handler is the function that will be called for each message received.
type Handler func(msg *pubsub.Message) error

// NewSubscription create a new subscription
func NewSubscription(subs *pubsub.Subscription, maxHandler int) Subscription {
	return Subscription{
		subs:       subs,
		MaxHandler: maxHandler,
	}
}

// NewPublish create a new publisher
func NewPublish(topic *pubsub.Topic) Publish {
	return Publish{
		topic: topic,
	}
}

// Send send a message to a topic
func (p Publish) Send(ctx context.Context, msg *pubsub.Message) error {

	err := p.topic.Send(ctx, msg)

	if err != nil {
		return err
	}

	return nil
}

// Receive receive messages from a subscription
func (s Subscription) Receive(ctx context.Context, handler Handler) error {
	if s.MaxHandler == 0 || s.MaxHandler < 0 {
		s.MaxHandler = 10
	}

	sem := make(chan struct{}, s.MaxHandler)
	for {
		msg, err := s.subs.Receive(ctx)
		if err != nil {
			return err
		}

		sem <- struct{}{} // acquire a semaphore slot

		go func() {
			defer func() { <-sem }() // release the semaphore
			err := handler(msg)
			if err != nil {
				if msg.Nackable() {
					msg.Nack()
				}
			}
			msg.Ack()
		}()
	}
}
