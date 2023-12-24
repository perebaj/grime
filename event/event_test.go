package event

import (
	"context"
	"testing"

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
