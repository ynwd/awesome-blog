package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ynwd/awesome-blog/pkg/module"
)

func (a *App) pubSubSubsribe(ctx context.Context) error {
	errChan := make(chan error, 1)

	go func() {
		topicSub := os.Getenv("GOOGLE_CLOUD_PUBSUB_SUBSCRIPTION")
		err := a.pubsub.Subscribe(ctx, topicSub, func(event interface{}) {

			// Convert to base event to check type
			data, _ := json.Marshal(event)
			var baseEvent module.BaseEvent
			if err := json.Unmarshal(data, &baseEvent); err != nil {
				log.Printf("Error unmarshaling event: %v", err)
				return
			}

			// Route to correct module based on event type
			for _, m := range a.modules {
				m.RegisterEventHandlers(ctx, baseEvent)
			}
		})
		if err != nil {
			errChan <- err
			return
		}
	}()

	// Wait for potential immediate subscription errors
	select {
	case err := <-errChan:
		return fmt.Errorf("failed to subscribe: %v", err)
	case <-time.After(time.Second):
		// Subscription started successfully
		return nil
	}
}
