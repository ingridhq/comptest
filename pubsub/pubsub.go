package pubsub

import (
	"context"
	"fmt"
	"log"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/golang/protobuf/proto"
)

func SetupPubsubTopic(ctx context.Context, project, topicID string, subIDs ...string) (*pubsub.Topic, error) {
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

func MustSetupPubSubTopic(ctx context.Context, project, topicID string, subIDs ...string) *pubsub.Topic {
	t, err := SetupPubsubTopic(ctx, project, topicID, subIDs...)
	if err != nil {
		log.Fatalf("Failed to setup pubsub for topic %q: %v", topicID, err)
	}
	return t
}

// MustSetupPubSubReceiver is used to create PubSub receiver.
// User must ensure that topic is created.
func MustSetupPubSubSubscription(ctx context.Context, projectID, topicID, subID string, cb func(context.Context, *pubsub.Message)) {
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

//MustPublishPubsubMsg is a helper function used to send custom PubSub message.
func MustPublishPubsubMsg(t *testing.T, ctx context.Context, s *pubsub.Topic, msg proto.Message) {
	res := s.Publish(ctx, &pubsub.Message{
		Data: []byte(msg.String()),
	})
	_, err := res.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to publish message to pubsub: %v", err)
	}
}
