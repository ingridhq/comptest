package comptest

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ingridhq/comptest"
	cppostgres "github.com/ingridhq/comptest/db/postgres"
	ctpubsub "github.com/ingridhq/comptest/pubsub"
	"github.com/ingridhq/comptest/waitfor"
	"github.com/kelseyhightower/envconfig"
)

var cfg struct {
	Port       string `envconfig:"PORT"`
	MetricPort string `envconfig:"METRICS_ADDR"`

	PubSubProjectID            string `envconfig:"PUBSUB_PROJECT_ID" default:"comptest-pubsub"`
	PubSubTopicReceived        string `envconfig:"PUBSUB_TOPIC_RECEIVED"`
	PubSubSubscriptionReceived string `envconfig:"PUBSUB_SUBSCRIPTION_RECEIVED"`

	PubSubTopicSend        string `envconfig:"PUBSUB_TOPIC_SEND"`
	PubSubSubscriptionSend string `envconfig:"PUBSUB_SUBSCRIPTION_SEND"`

	DBPostgresDSN string `envconfig:"DB_POSTGRES_DSN"`
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

	envconfig.MustProcess("", &cfg)

	// Initialize comptest lib.
	c := comptest.New()

	postgresDB := cppostgres.Database(cfg.DBPostgresDSN)

	c.HealthChecks(
		postgresDB,
		waitfor.TCP(os.Getenv("PUBSUB_EMULATOR_HOST")),
	)

	// Setting up all dependencies needed in tests...
	if err := postgresDB.CreateDatabase(context.Background()); err != nil {
		log.Fatalf("could not create database: %v", err)
	}

	if err := postgresDB.RunDownMigrations("file://../migrations/"); err != nil {
		log.Fatalf("Failed to run down migration: %v", err)
	}
	if err := postgresDB.RunUpMigrations("file://../migrations/"); err != nil {
		log.Fatalf("Failed to run up migration: %v", err)
	}

	sender := ctpubsub.MustSetupTopic(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicReceived, cfg.PubSubSubscriptionReceived)
	ctpubsub.MustSetupTopic(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicSend, cfg.PubSubSubscriptionSend)

	receiver := make(chan *pubsub.Message)
	ctpubsub.MustSetupSubscription(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicSend, cfg.PubSubSubscriptionSend,
		func(ctx context.Context, message *pubsub.Message) {
			receiver <- message
		},
	)

	// Build, run, wait for service and run tests...
	cleanup := c.BuildAndRun("../main.go", waitfor.HTTP(fmt.Sprintf("http://%s/readiness", cfg.MetricPort)))
	defer cleanup()

	env = Environment{
		Sender:   sender,
		Receiver: receiver,
		// or connection to freshly ran service.
	}

	t.Run()
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
