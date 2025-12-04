package domain_test

import (
	"nostar/internal/relay/domain"
	"testing"

	"github.com/nbd-wtf/go-nostr"
)

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

func TestEvent_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   domain.Event
		wantErr bool
	}{
		{
			name: "valid event",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "valid-pubkey",
				Signature: "valid-signature",
				CreatedAt: 1234567890,
				Kind:      1,
				Tags:      [][]string{},
				Content:   "test content",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			event: domain.Event{
				ID:        "",
				PubKey:    "valid-pubkey",
				Signature: "valid-signature",
				CreatedAt: 1234567890,
				Kind:      1,
			},
			wantErr: true,
		},
		{
			name: "empty PubKey",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "",
				Signature: "valid-signature",
				CreatedAt: 1234567890,
				Kind:      1,
			},
			wantErr: true,
		},
		{
			name: "empty Signature",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "valid-pubkey",
				Signature: "",
				CreatedAt: 1234567890,
				Kind:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid CreatedAt (zero)",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "valid-pubkey",
				Signature: "valid-signature",
				CreatedAt: 0,
				Kind:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid CreatedAt (negative)",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "valid-pubkey",
				Signature: "valid-signature",
				CreatedAt: -1,
				Kind:      1,
			},
			wantErr: true,
		},
		{
			name: "invalid Kind (negative)",
			event: domain.Event{
				ID:        "valid-id",
				PubKey:    "valid-pubkey",
				Signature: "valid-signature",
				CreatedAt: 1234567890,
				Kind:      -1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := tt.event.Validate()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Validate() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Validate() succeeded unexpectedly")
			}
		})
	}
}

func TestEvent_CheckSignature(t *testing.T) {
	// Generate a valid event for testing
	validEvent := createValidTestEvent("hello world", 1)

	tests := []struct {
		name    string
		event   domain.Event
		want    bool
		wantErr bool
	}{
		{
			name:    "valid signature",
			event:   validEvent,
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid signature (wrong signature)",
			event: domain.Event{
				ID:        validEvent.ID,
				PubKey:    validEvent.PubKey,
				Signature: "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				CreatedAt: validEvent.CreatedAt,
				Kind:      validEvent.Kind,
				Tags:      [][]string{},
				Content:   validEvent.Content,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid ID (wrong ID)",
			event: domain.Event{
				ID:        "0000000000000000000000000000000000000000000000000000000000000000",
				PubKey:    validEvent.PubKey,
				Signature: validEvent.Signature,
				CreatedAt: validEvent.CreatedAt,
				Kind:      validEvent.Kind,
				Tags:      [][]string{},
				Content:   validEvent.Content,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid signature format",
			event: domain.Event{
				ID:        validEvent.ID,
				PubKey:    validEvent.PubKey,
				Signature: "invalid-signature-format",
				CreatedAt: validEvent.CreatedAt,
				Kind:      validEvent.Kind,
				Tags:      [][]string{},
				Content:   validEvent.Content,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "tampered content",
			event: domain.Event{
				ID:        validEvent.ID,
				PubKey:    validEvent.PubKey,
				Signature: validEvent.Signature,
				CreatedAt: validEvent.CreatedAt,
				Kind:      validEvent.Kind,
				Tags:      [][]string{},
				Content:   "tampered content", // Changed content
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := tt.event.CheckSignature()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CheckSignature() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CheckSignature() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("CheckSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
