package domain

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
