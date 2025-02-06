package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type PubSubClient interface {
	Publish(ctx context.Context, data interface{}) error
	Subscribe(ctx context.Context, subscriptionID string, handler func(event interface{})) error
	Close()
}

type pubSubClient struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

func NewPubSubClient(projectID, topicID string) (PubSubClient, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile("serviceAccountKey.json")
	client, err := pubsub.NewClient(ctx, projectID, opt)
	if err != nil {
		return nil, err
	}

	topic := client.Topic(topicID)
	return &pubSubClient{
		client: client,
		topic:  topic,
	}, nil
}

func (p *pubSubClient) Publish(ctx context.Context, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	log.Printf("Publishing JSON data: %s", string(jsonData))

	result := p.topic.Publish(ctx, &pubsub.Message{
		Data: jsonData,
	})

	_, err = result.Get(ctx)
	return err
}

func (p *pubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler func(event interface{})) error {
	sub := p.client.Subscription(subscriptionID)

	// Check if subscription exists
	exists, err := sub.Exists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check subscription: %w", err)
	}

	if !exists {
		// Create new subscription if not exists
		sub, err = p.client.CreateSubscription(ctx, subscriptionID, pubsub.SubscriptionConfig{
			Topic:            p.topic,
			AckDeadline:      20 * time.Second,
			ExpirationPolicy: 25 * time.Hour,
		})
		if err != nil {
			return fmt.Errorf("failed to create subscription: %w", err)
		}
	}

	// Start receiving messages
	return sub.Receive(ctx, func(msgCtx context.Context, msg *pubsub.Message) {
		var event interface{}

		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			msg.Nack()
			return
		}

		// Execute handler
		handler(event)
		msg.Ack()
	})
}

func (p *pubSubClient) Close() {
	if p.topic != nil {
		p.topic.Stop()
	}
	if p.client != nil {
		p.client.Close()
	}
}
