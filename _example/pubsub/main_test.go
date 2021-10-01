package comptest

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"

	ctpubsub "github.com/ingridhq/comptest/pubsub"
)

type config struct {
	PubSubProjectID    string `envconfig:"PUBSUB_PROJECT_ID" default:"comptest-pubsub"`
	PubSubTopic        string `envconfig:"PUBSUB_TOPIC"`
	PubSubSubscription string `envconfig:"PUBSUB_SUBSCRIPTION"`
}

type Environment struct {
	Sender   *pubsub.Topic
	Receiver chan *pubsub.Message
}

var env Environment

func TestMain(t *testing.M) {
	if os.Getenv("RUN_COMPONENT_TESTS") != "true" {
		return
	}

	var cfg config
	envconfig.MustProcess("", &cfg)

	ctx := context.Background()
	sender := ctpubsub.MustSetupPubSubTopic(ctx, cfg.PubSubProjectID, cfg.PubSubTopic, cfg.PubSubSubscription)

	receiver := make(chan *pubsub.Message)
	ctpubsub.MustSetupPubSubSubscription(
		ctx,
		cfg.PubSubProjectID,
		cfg.PubSubTopic,
		cfg.PubSubSubscription,
		func(ctx context.Context, message *pubsub.Message) {
			receiver <- message
		})

	env = Environment{
		Sender:   sender,
		Receiver: receiver,
	}
	t.Run()
}

func TestPubSub(t *testing.T) {
	message := &pubsub.Message{
		Data: []byte("Test"),
	}

	env.Sender.Publish(context.Background(), message)

	data := <-env.Receiver

	if dr := string(data.Data); dr != string(message.Data) {
		t.Errorf("wrong data, exp %q, got %q", string(message.Data), dr)
	}
}
