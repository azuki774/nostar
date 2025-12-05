package domain

// Subscription represents a REQ subscription with filters.
// Filter matching logic can be added here.
type Subscription struct {
	ID      string   // subscription ID from client
	Filters []Filter // filtering events
}

// Matches returns whether the event satisfies the subscription filters.
func (s Subscription) Matches(evt Event) bool {
	// 引数のイベントが、既存のサブスクリプションの FIlters 条件に合うかを検索
	// Filter 同士は、OR 条件なのでマッチすれば true
	for _, filter := range s.Filters {
		if filter.Matches(evt) {
			return true
		}
	}
	return false
}
