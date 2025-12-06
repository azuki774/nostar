package usecase

import "nostar/internal/relay/domain"

// EventMessage wraps an EVENT message from a client.
type EventMessage struct {
	Event domain.Event
}

// ReqMessage wraps a REQ with filters.
type ReqMessage struct {
	ConnectionID domain.ConnectionID
	Subscription domain.Subscription
}

// CloseMessage represents a CLOSE request for a subscription ID.
type CloseMessage struct {
	ConnectionID   domain.ConnectionID
	SubscriptionID string
}
