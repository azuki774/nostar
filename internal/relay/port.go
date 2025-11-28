package relay

import (
	"context"

	"nostar/internal/relay/domain"
)

// EventStore persists and queries Nostr events.
type EventStore interface {
	Save(ctx context.Context, evt domain.Event) error
	Query(ctx context.Context, sub domain.Subscription) ([]domain.Event, error)
}
