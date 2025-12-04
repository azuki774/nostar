package usecase

import (
	"context"

	"nostar/internal/relay"
	"nostar/internal/relay/domain"
)

// RelayService bundles the Nostr relay business use cases.
// It does not know about transports (WebSocket/HTTP); those call into this type.
type RelayService struct {
	store relay.EventStore
}

func NewRelayService(store relay.EventStore) *RelayService {
	return &RelayService{
		store: store,
	}
}

// HandleEvent processes an EVENT message: validation, persistence, and fanout.
func (s *RelayService) HandleEvent(ctx context.Context, msg EventMessage) error {
	// Validate event fields
	if err := msg.Event.Validate(); err != nil {
		return err
	}

	// Verify signature and ID
	valid, err := msg.Event.CheckSignature()
	if err != nil {
		return err
	}
	if !valid {
		return err
	}

	// Save to store
	if err := s.store.Save(ctx, msg.Event); err != nil {
		return err
	}
	return nil
}

// HandleReq processes a REQ: query stored events and start live subscription if available.
func (s *RelayService) HandleReq(ctx context.Context, msg ReqMessage) ([]domain.Event, error) {
	events, err := s.store.Query(ctx, msg.Subscription)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// HandleClose processes CLOSE; any subscription cleanup would happen here.
func (s *RelayService) HandleClose(ctx context.Context, msg CloseMessage) error {
	return nil
}
