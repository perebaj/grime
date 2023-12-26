package event

import (
	"context"
	"testing"
	"time"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

func TestSubscription_Receive(t *testing.T) {
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

	s := NewSubscription(subs, 1)

	go func() {
		err := s.Receive(ctx, run)
		if err != nil {
			t.Error(err)
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

	/*
		Here we have and infinite loop case, so what I'm doing is to wait for
		a determined time and then exit the test, without throwing an error, just to
		avoid that the event loop will block the test execution.
	*/
	<-time.After(100 * time.Millisecond)
}

func newTopic(t *testing.T) string {
	return "mem://" + t.Name()
}
