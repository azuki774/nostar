package domain

// Subscription represents a REQ subscription with filters.
// Filter matching logic can be added here.
type Subscription struct {
	ID      string   // subscription ID from client
	Authors []string // optional pubkeys
	Kinds   []int    // optional kinds
	Tags    [][]string
	Since   *int64
	Until   *int64
	Limit   *int
}

// Matches returns whether the event satisfies the subscription filters.
// TODO: implement proper filtering per Nostr spec.
func (s Subscription) Matches(Event) bool {
	return true
}
