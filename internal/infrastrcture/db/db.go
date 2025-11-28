package db

import (
	"context"
	"nostar/internal/relay/domain"

	"go.uber.org/zap"
)

type EventStore struct {
}

// TODO: 具体的に実装する
func NewEventStore() *EventStore {
	return &EventStore{}
}

func (e *EventStore) Save(ctx context.Context, evt domain.Event) error {
	zap.S().Infow("not yet implemeneted", "type", "Save")
	return nil
}

func (e *EventStore) Query(ctx context.Context, sub domain.Subscription) ([]domain.Event, error) {
	zap.S().Infow("not yet implemeneted", "type", "Query")
	return []domain.Event{}, nil
}
