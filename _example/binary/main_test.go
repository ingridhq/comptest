package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"

	"github.com/ingridhq/comptest"
	ctpubsub "github.com/ingridhq/comptest/pubsub"
	"github.com/ingridhq/comptest/waitfor"
)

var cfg Config

type Environment struct {
	Sender   *pubsub.Topic
	Receiver chan *pubsub.Message
}

var env Environment

func TestMain(t *testing.M) {
	if os.Getenv("RUN_COMPONENT_TESTS") != "true" {
		return
	}

	c := comptest.New(t)
	c.HealthChecks(
		waitfor.TCP(os.Getenv("PUBSUB_EMULATOR_HOST")),
	)

	envconfig.MustProcess("", &cfg)

	sender := ctpubsub.MustSetupTopic(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicReceived, cfg.PubSubSubscriptionReceived)
	ctpubsub.MustSetupTopic(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicSend, cfg.PubSubSubscriptionSend)

	receiver := make(chan *pubsub.Message)
	ctpubsub.MustSetupSubscription(
		context.Background(),
		cfg.PubSubProjectID,
		cfg.PubSubTopicSend,
		cfg.PubSubSubscriptionSend,
		func(ctx context.Context, message *pubsub.Message) {
			receiver <- message
		},
	)

	env = Environment{
		Sender:   sender,
		Receiver: receiver,
	}

	c.BuildAndRun("main.go", waitfor.HTTP(fmt.Sprintf("http://%s/readiness", cfg.MetricPort)))
}

func Test_HTTP(t *testing.T) {
	resp, err := http.Get(fmt.Sprintf("http://%v/", cfg.Port))
	if err != nil {
		t.Fatalf("could not do get request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	if string(body) != "Potato" {
		t.Fatalf("unexpected response: %v", string(body))
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code : %v", resp.StatusCode)
	}
}

func Test_PubSubRecivied(t *testing.T) {
	_, err := http.Get(fmt.Sprintf("http://%v/event", cfg.Port))
	if err != nil {
		t.Fatalf("could not do get request: %v", err)
	}

	data := <-env.Receiver

	if dr := string(data.Data); dr != "Test" {
		t.Errorf("wrong data, exp %q, got %q", "Test", dr)
	}
}

func Test_PubSubSend(t *testing.T) {
	ctx := context.Background()
	env.Sender.Publish(ctx, &pubsub.Message{
		Data: []byte("empty message"),
	})

	time.Sleep(250 * time.Millisecond)
	resp, err := http.Get(fmt.Sprintf("http://%v/event_count", cfg.Port))
	if err != nil {
		t.Fatalf("could not do get request: %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	if string(body) != "Current count: 1" {
		t.Fatalf("unexpected response: %v", string(body))
	}
}
