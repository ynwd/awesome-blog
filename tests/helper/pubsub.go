package helper

import "context"

type MockPubSub struct {
	PublishFunc   func(ctx context.Context, event interface{}) error
	SubscribeFunc func(ctx context.Context, subscriptionID string, handler func(event interface{})) error
}

func (m *MockPubSub) Publish(ctx context.Context, event interface{}) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(ctx, event)
	}
	return nil
}

func (m *MockPubSub) Subscribe(ctx context.Context, subscriptionID string, handler func(event interface{})) error {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(ctx, subscriptionID, handler)
	}
	return nil
}

func (m *MockPubSub) Close() {}
