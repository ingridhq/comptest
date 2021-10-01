package pubsub

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

func newSubscription(ctx context.Context, projectID, topicID, subscriptionID string) (*pubsub.Subscription, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create client for project %q: %v", projectID, err)
	}

	topic := client.Topic(topicID)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("topic %q does not exist", topicID)
	}

	subscription, err := ensureSubscription(ctx, client, topic, subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to subscription %q: %w", subscriptionID, err)
	}

	return subscription, nil
}

func ensureTopic(ctx context.Context, cli *pubsub.Client, topicID string) (*pubsub.Topic, error) {
	t := cli.Topic(topicID)
	ok, err := t.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if topic exists: %w", err)
	}
	if ok {
		return t, nil
	}
	t, err = cli.CreateTopic(ctx, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to create topic %q: %w", topicID, err)
	}
	return t, nil
}

func ensureSubscription(ctx context.Context, cli *pubsub.Client, t *pubsub.Topic, subID string) (*pubsub.Subscription, error) {
	s := cli.Subscription(subID)
	ok, err := s.Exists(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if subscription %q exists: %w", subID, err)
	}
	if ok {
		return s, nil
	}
	_, err = cli.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{
		Topic: t,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription %q: %w", subID, err)
	}
	return s, nil
}
