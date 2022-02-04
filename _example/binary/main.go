package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/kelseyhightower/envconfig"

	ctpubsub "github.com/ingridhq/comptest/pubsub"
)

var count int64

type Config struct {
	Port       string `envconfig:"PORT"`
	MetricPort string `envconfig:"METRICS_ADDR"`

	PubSubProjectID            string `envconfig:"PUBSUB_PROJECT_ID" default:"comptest-pubsub"`
	PubSubTopicReceived        string `envconfig:"PUBSUB_TOPIC_RECEIVED"`
	PubSubSubscriptionReceived string `envconfig:"PUBSUB_SUBSCRIPTION_RECEIVED"`

	PubSubTopicSend        string `envconfig:"PUBSUB_TOPIC_SEND"`
	PubSubSubscriptionSend string `envconfig:"PUBSUB_SUBSCRIPTION_SEND"`
}

func main() {
	log.Println("START")
	var cfg Config
	envconfig.MustProcess("", &cfg)

	sender := ctpubsub.MustSetupTopic(context.Background(), cfg.PubSubProjectID, cfg.PubSubTopicSend, cfg.PubSubSubscriptionSend)

	ctpubsub.MustSetupSubscription(
		context.Background(),
		cfg.PubSubProjectID,
		cfg.PubSubTopicReceived,
		cfg.PubSubSubscriptionReceived,
		func(ctx context.Context, message *pubsub.Message) {
			atomic.AddInt64(&count, 1)

			message.Ack()
		},
	)

	mainMux := http.NewServeMux()

	mainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Potato endpoint")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Potato"))
	})

	mainMux.HandleFunc("/event", func(w http.ResponseWriter, r *http.Request) {
		message := &pubsub.Message{
			Data: []byte("Test"),
		}

		sender.Publish(context.Background(), message)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ok"))
	})

	mainMux.HandleFunc("/event_count", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("Current count: %v", atomic.LoadInt64(&count))))
	})

	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/readiness", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Readiness endpoint")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready!!!"))
	})

	go func() {
		// Sleep to simulate system is not ready to use
		time.Sleep(time.Second * 1)
		log.Println("Start metrics server")
		if err := http.ListenAndServe(cfg.MetricPort, metricsMux); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Start main server")
	if err := http.ListenAndServe(cfg.Port, mainMux); err != nil {
		log.Fatal(err)
	}
}
