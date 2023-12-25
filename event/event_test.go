package event

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

func TestEvent(t *testing.T) {
	ctx := context.Background()
	topic, err := pubsub.OpenTopic(ctx, "mem://topicA")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = topic.Shutdown(ctx)
	})

	// Open a *pubsub.Subscription that receives from the topic.
	subs, err := pubsub.OpenSubscription(ctx, "mem://topicA")
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	msgCh := make(chan *pubsub.Message)
	go func() {
		for {
			msg, err := subs.Receive(ctx)
			if err != nil {
				t.Error(err)
				break
			}
			msg.Ack()
			msgCh <- msg
			close(done)
		}
	}()

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte("Hello world!"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if msg := <-msgCh; string(msg.Body) != "Hello world!" {
		t.Errorf("got %s, want %s", string(msg.Body), "Hello world!")
	}

	<-done
}

func TestEvent_Send(t *testing.T) {
	ctx := context.Background()
	topicName := newTopic(t)
	topic, err := pubsub.OpenTopic(ctx, topicName)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = topic.Shutdown(ctx)
	})

	msgCh := make(chan *pubsub.Message)
	run := func(msg *pubsub.Message) error {
		msgCh <- msg
		return nil
	}

	subs, err := pubsub.OpenSubscription(ctx, topicName)
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		err := subscribe(run, 1, subs, ctx)
		if err != nil {
			t.Error(err)
		}
		close(done)
	}()

	err = topic.Send(ctx, &pubsub.Message{
		Body: []byte("Hello world!"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if msg := <-msgCh; string(msg.Body) != "Hello world!" {
		fmt.Println(string(msg.Body))
		t.Errorf("got %s, want %s", string(msg.Body), "Hello world!")
	}

	select {
	case <-done:
	/*
		Here we have and infinite loop case, so what I'm doing is to wait for
		a determined time and then exit the test, without throwing an error, just to
		avoid that the event loop will block the test execution.
	*/
	case <-time.After(1 * time.Second):
	}
}

func subscribe(run func(msg *pubsub.Message) error, maxHandler int, subs *pubsub.Subscription, ctx context.Context) error {
	sem := make(chan struct{}, maxHandler)
	for {
		msg, err := subs.Receive(ctx)
		if err != nil {
			return err
		}

		sem <- struct{}{}

		go func() {
			defer func() { <-sem }() // release the semaphore
			err = run(msg)
			if err != nil {
				if msg.Nackable() {
					msg.Nack()
				}
				return
			}
			msg.Ack()
		}()
	}
}

func newTopic(t *testing.T) string {
	return "mem://" + t.Name()
}
