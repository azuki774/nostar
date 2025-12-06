package usecase_test

import (
	"context"
	"errors"
	"nostar/internal/relay"
	"nostar/internal/relay/domain"
	"nostar/internal/relay/usecase"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

// mockEventStore is a mock implementation of relay.EventStore for testing
type mockEventStore struct {
	saveFunc  func(ctx context.Context, evt domain.Event) error
	queryFunc func(ctx context.Context, sub domain.Subscription) ([]domain.Event, error)
	saveCalls int // Track number of times Save was called
}

func (m *mockEventStore) Save(ctx context.Context, evt domain.Event) error {
	m.saveCalls++
	if m.saveFunc != nil {
		return m.saveFunc(ctx, evt)
	}
	return nil
}

func (m *mockEventStore) Query(ctx context.Context, sub domain.Subscription) ([]domain.Event, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sub)
	}
	return nil, nil
}

// createValidTestEvent creates a valid Nostr event for testing
func createValidTestEvent(content string, kind int) domain.Event {
	// Generate a test key pair
	sk := nostr.GeneratePrivateKey()
	pk, _ := nostr.GetPublicKey(sk)

	// Create a nostr event
	nostrEvent := nostr.Event{
		PubKey:    pk,
		CreatedAt: nostr.Timestamp(1671028937),
		Kind:      kind,
		Tags:      nostr.Tags{},
		Content:   content,
	}

	// Sign the event
	nostrEvent.Sign(sk)

	// Convert to domain.Event
	return domain.Event{
		ID:        nostrEvent.ID,
		PubKey:    nostrEvent.PubKey,
		Signature: nostrEvent.Sig,
		CreatedAt: int64(nostrEvent.CreatedAt),
		Kind:      nostrEvent.Kind,
		Tags:      [][]string{},
		Content:   nostrEvent.Content,
	}
}

func TestRelayService_HandleEvent(t *testing.T) {
	validEvent := createValidTestEvent("test content", 1)

	tests := []struct {
		name      string
		store     relay.EventStore
		msg       usecase.EventMessage
		wantErr   bool
		checkSave bool // Whether Save should have been called
	}{
		{
			name: "valid event is saved successfully",
			store: &mockEventStore{
				saveFunc: func(ctx context.Context, evt domain.Event) error {
					return nil
				},
			},
			msg: usecase.EventMessage{
				Event: validEvent,
			},
			wantErr:   false,
			checkSave: true,
		},
		{
			name: "validation error - empty ID",
			store: &mockEventStore{
				saveFunc: func(ctx context.Context, evt domain.Event) error {
					t.Error("Save should not be called when validation fails")
					return nil
				},
			},
			msg: usecase.EventMessage{
				Event: domain.Event{
					ID:        "", // Invalid: empty ID
					PubKey:    "valid-pubkey",
					Signature: "valid-signature",
					CreatedAt: 1234567890,
					Kind:      1,
				},
			},
			wantErr:   true,
			checkSave: false,
		},
		{
			name: "signature verification error - wrong signature",
			store: &mockEventStore{
				saveFunc: func(ctx context.Context, evt domain.Event) error {
					t.Error("Save should not be called when signature verification fails")
					return nil
				},
			},
			msg: usecase.EventMessage{
				Event: domain.Event{
					ID:        validEvent.ID,
					PubKey:    validEvent.PubKey,
					Signature: "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
					CreatedAt: validEvent.CreatedAt,
					Kind:      validEvent.Kind,
					Tags:      [][]string{},
					Content:   validEvent.Content,
				},
			},
			wantErr:   true,
			checkSave: false,
		},
		{
			name: "store save error",
			store: &mockEventStore{
				saveFunc: func(ctx context.Context, evt domain.Event) error {
					return errors.New("database error")
				},
			},
			msg: usecase.EventMessage{
				Event: validEvent,
			},
			wantErr:   true,
			checkSave: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connPool := domain.NewConnectionPool()
			s := usecase.NewRelayService(tt.store, connPool)
			mock, ok := tt.store.(*mockEventStore)
			if ok {
				mock.saveCalls = 0 // Reset counter
			}

			gotErr := s.HandleEvent(context.Background(), tt.msg)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("HandleEvent() failed: %v", gotErr)
				}
			} else {
				if tt.wantErr {
					t.Fatal("HandleEvent() succeeded unexpectedly")
				}
			}

			// Check if Save was called when expected
			if ok && tt.checkSave && mock.saveCalls != 1 {
				t.Errorf("Expected Save to be called once, but was called %d times", mock.saveCalls)
			}
			if ok && !tt.checkSave && mock.saveCalls != 0 {
				t.Errorf("Expected Save not to be called, but was called %d times", mock.saveCalls)
			}
		})
	}
}
