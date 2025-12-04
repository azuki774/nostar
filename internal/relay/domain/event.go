package domain

import (
	"fmt"

	"github.com/nbd-wtf/go-nostr"
)

// Event represents a Nostr event on the wire.
// Fields are minimal; validation/verification should live alongside this struct.
type Event struct {
	ID        string     `json:"id"`
	PubKey    string     `json:"pubkey"`
	Signature string     `json:"sig"`
	CreatedAt int64      `json:"created_at"`
	Kind      int        `json:"kind"`
	Tags      [][]string `json:"tags"`
	Content   string     `json:"content"`
	Raw       []byte     `json:"-"` // raw message for hashing/verification
}

// Validate performs basic validation on the event fields.
func (e *Event) Validate() error {
	if e.ID == "" {
		return fmt.Errorf("event ID is empty")
	}
	if e.PubKey == "" {
		return fmt.Errorf("event pubkey is empty")
	}
	if e.Signature == "" {
		return fmt.Errorf("event signature is empty")
	}
	if e.CreatedAt <= 0 {
		return fmt.Errorf("event created_at is invalid")
	}
	if e.Kind < 0 {
		return fmt.Errorf("event kind is invalid")
	}
	return nil
}

// CheckSignature verifies the event's ID and signature using go-nostr.
// 署名周りは独自実装のリスクが高いので、ライブラリを使う
func (e *Event) CheckSignature() (bool, error) {
	// Convert tags to go-nostr Tags type
	tags := make(nostr.Tags, len(e.Tags))
	for i, tag := range e.Tags {
		tags[i] = nostr.Tag(tag)
	}

	// Convert to go-nostr Event
	nostrEvent := &nostr.Event{
		ID:        e.ID,
		PubKey:    e.PubKey,
		Sig:       e.Signature,
		CreatedAt: nostr.Timestamp(e.CreatedAt),
		Kind:      e.Kind,
		Tags:      tags,
		Content:   e.Content,
	}

	// Verify ID (computed hash matches the provided ID)
	if !nostrEvent.CheckID() {
		return false, fmt.Errorf("event ID does not match computed hash")
	}

	// Verify signature (schnorr signature is valid for the pubkey)
	ok, err := nostrEvent.CheckSignature()
	if err != nil {
		return false, fmt.Errorf("signature verification failed: %w", err)
	}
	if !ok {
		return false, fmt.Errorf("invalid signature")
	}

	return true, nil
}
