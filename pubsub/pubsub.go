package pubsub

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub"
)

func SetupTopic(ctx context.Context, project, topicID string, subIDs ...string) (*pubsub.Topic, error) {
	c, err := pubsub.NewClient(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}
	topic, err := ensureTopic(ctx, c, topicID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure that topic %q exists: %w", topicID, err)
	}
	for _, sub := range subIDs {
		_, err := ensureSubscription(ctx, c, topic, sub)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscription %q: %w", sub, err)
		}
	}
	return topic, nil
}

func MustSetupTopic(ctx context.Context, project, topicID string, subIDs ...string) *pubsub.Topic {
	t, err := SetupTopic(ctx, project, topicID, subIDs...)
	if err != nil {
		log.Fatalf("Failed to setup pubsub for topic %q: %v", topicID, err)
	}
	return t
}

// MustSetupSubscription is used to create PubSub receiver.
// User must ensure that topic is created.
func MustSetupSubscription(ctx context.Context, projectID, topicID, subID string, cb func(context.Context, *pubsub.Message)) {
	r, err := newSubscription(ctx, projectID, topicID, subID)
	if err != nil {
		log.Fatalf("Failed to create receiver for topic %q: %v", topicID, err)
	}

	go func() {
		err := r.Receive(ctx, cb)
		if err != nil {
			log.Fatalf("Failed to start receiver for topic %q: %v", topicID, err)
		}
	}()
}
