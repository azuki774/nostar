package memory_test

import (
	"nostar/internal/infrastructure/memory"
	"nostar/internal/relay/domain"
	"testing"
)

func TestMemorySubscriptionRegistry_Register(t *testing.T) {
	tests := []struct {
		name    string
		connID  domain.ConnectionID
		sub     domain.Subscription
		wantErr bool
	}{
		{
			name:   "successful subscription registration",
			connID: "test-conn-1",
			sub: domain.Subscription{
				ID: "sub-1",
				Filters: []domain.Filter{{
					Kinds: []int{1},
				}},
			},
			wantErr: false,
		},
		{
			name:   "multiple subscriptions for same connection ID",
			connID: "test-conn-1",
			sub: domain.Subscription{
				ID: "sub-2",
				Filters: []domain.Filter{{
					Kinds: []int{2},
				}},
			},
			wantErr: false,
		},
		{
			name:   "subscription with same name on different connection",
			connID: "test-conn-2",
			sub: domain.Subscription{
				ID: "sub-1",
				Filters: []domain.Filter{{
					Kinds: []int{1},
				}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msr := memory.NewMemorySubscriptionRegistry()
			gotErr := msr.Register(tt.connID, tt.sub)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", gotErr, tt.wantErr)
				return
			}
		})
	}
}

func TestMemorySubscriptionRegistry_Unregister(t *testing.T) {
	tests := []struct {
		name    string
		connID  domain.ConnectionID
		subID   string
		setup   func(*memory.MemorySubscriptionRegistry) // テスト前のセットアップ
		wantErr bool
	}{
		{
			name:   "unregister existing subscription",
			connID: "test-conn-1",
			subID:  "sub-1",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{{Kinds: []int{1}}},
				}
				msr.Register("test-conn-1", sub)
			},
			wantErr: false,
		},
		{
			name:   "unregister non-existent subscription",
			connID: "test-conn-1",
			subID:  "non-existent",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{{Kinds: []int{1}}},
				}
				msr.Register("test-conn-1", sub)
			},
			wantErr: false,
		},
		{
			name:    "unregister from non-existent connection",
			connID:  "non-existent-conn",
			subID:   "sub-1",
			setup:   func(*memory.MemorySubscriptionRegistry) {},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msr := memory.NewMemorySubscriptionRegistry()
			tt.setup(msr.(*memory.MemorySubscriptionRegistry))
			gotErr := msr.Unregister(tt.connID, tt.subID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("Unregister() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
		})
	}
}

func TestMemorySubscriptionRegistry_UnregisterAll(t *testing.T) {
	tests := []struct {
		name    string
		connID  domain.ConnectionID
		setup   func(*memory.MemorySubscriptionRegistry)
		verify  func(*testing.T, *memory.MemorySubscriptionRegistry, domain.ConnectionID)
		wantErr bool
	}{
		{
			name:   "unregister all subscriptions and verify removal",
			connID: "test-conn-1",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub1 := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				sub2 := domain.Subscription{ID: "sub-2", Filters: []domain.Filter{{Kinds: []int{2}}}}
				msr.Register("test-conn-1", sub1)
				msr.Register("test-conn-1", sub2)
			},
			verify: func(t *testing.T, msr *memory.MemorySubscriptionRegistry, connID domain.ConnectionID) {
				// UnregisterAll 後にイベントを送信してもマッチしないことを確認
				event := domain.Event{Kind: 1}
				matching := msr.FindMatchingConnections(event)
				if len(matching) != 0 {
					t.Errorf("Expected no matching connections after UnregisterAll, got %v", matching)
				}
				// 別のイベント（kind=2）も確認
				event2 := domain.Event{Kind: 2}
				matching2 := msr.FindMatchingConnections(event2)
				if len(matching2) != 0 {
					t.Errorf("Expected no matching connections after UnregisterAll for kind=2, got %v", matching2)
				}
			},
			wantErr: false,
		},
		{
			name:    "unregister all from non-existent connection",
			connID:  "non-existent-conn",
			setup:   func(*memory.MemorySubscriptionRegistry) {},
			verify:  func(*testing.T, *memory.MemorySubscriptionRegistry, domain.ConnectionID) {},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msr := memory.NewMemorySubscriptionRegistry()
			tt.setup(msr.(*memory.MemorySubscriptionRegistry))
			gotErr := msr.UnregisterAll(tt.connID)
			if (gotErr != nil) != tt.wantErr {
				t.Errorf("UnregisterAll() error = %v, wantErr %v", gotErr, tt.wantErr)
			}
			// 内部状態の検証
			tt.verify(t, msr.(*memory.MemorySubscriptionRegistry), tt.connID)
		})
	}
}

func TestMemorySubscriptionRegistry_FindMatchingSubscriptions(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*memory.MemorySubscriptionRegistry)
		event domain.Event
		want  []domain.SubscriptionMatch
	}{
		{
			name: "single subscription matches event",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				msr.Register("conn-1", sub)
			},
			event: domain.Event{Kind: 1},
			want: []domain.SubscriptionMatch{
				{ConnectionID: "conn-1", SubscriptionID: "sub-1"},
			},
		},
		{
			name: "multiple subscriptions match same event",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub1 := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				sub2 := domain.Subscription{ID: "sub-2", Filters: []domain.Filter{{Kinds: []int{1}}}}
				msr.Register("conn-1", sub1)
				msr.Register("conn-1", sub2)
			},
			event: domain.Event{Kind: 1},
			want: []domain.SubscriptionMatch{
				{ConnectionID: "conn-1", SubscriptionID: "sub-1"},
				{ConnectionID: "conn-1", SubscriptionID: "sub-2"},
			},
		},
		{
			name: "multiple connections have matching subscriptions",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub1 := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				sub2 := domain.Subscription{ID: "sub-2", Filters: []domain.Filter{{Kinds: []int{1}}}}
				msr.Register("conn-1", sub1)
				msr.Register("conn-2", sub2)
			},
			event: domain.Event{Kind: 1},
			want: []domain.SubscriptionMatch{
				{ConnectionID: "conn-1", SubscriptionID: "sub-1"},
				{ConnectionID: "conn-2", SubscriptionID: "sub-2"},
			},
		},
		{
			name: "no subscriptions match event",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				msr.Register("conn-1", sub)
			},
			event: domain.Event{Kind: 2},
			want:  []domain.SubscriptionMatch{},
		},
		{
			name: "connection has multiple subscriptions but only one matches",
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub1 := domain.Subscription{ID: "sub-1", Filters: []domain.Filter{{Kinds: []int{1}}}}
				sub2 := domain.Subscription{ID: "sub-2", Filters: []domain.Filter{{Kinds: []int{2}}}}
				msr.Register("conn-1", sub1)
				msr.Register("conn-1", sub2)
			},
			event: domain.Event{Kind: 1},
			want: []domain.SubscriptionMatch{
				{ConnectionID: "conn-1", SubscriptionID: "sub-1"},
			},
		},
		{
			name:  "empty registry returns no matches",
			setup: func(msr *memory.MemorySubscriptionRegistry) {},
			event: domain.Event{Kind: 1},
			want:  []domain.SubscriptionMatch{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msr := memory.NewMemorySubscriptionRegistry()
			tt.setup(msr.(*memory.MemorySubscriptionRegistry))
			got := msr.FindMatchingSubscriptions(tt.event)
			if !equalSubscriptionMatches(got, tt.want) {
				t.Errorf("FindMatchingSubscriptions() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ヘルパー関数：SubscriptionMatchスライスの比較（順序無視）
func equalSubscriptionMatches(a, b []domain.SubscriptionMatch) bool {
	if len(a) != len(b) {
		return false
	}

	// マップを使ってカウント
	count := make(map[string]int)
	for _, match := range a {
		key := string(match.ConnectionID) + ":" + match.SubscriptionID
		count[key]++
	}
	for _, match := range b {
		key := string(match.ConnectionID) + ":" + match.SubscriptionID
		count[key]--
	}

	for _, c := range count {
		if c != 0 {
			return false
		}
	}
	return true
}
