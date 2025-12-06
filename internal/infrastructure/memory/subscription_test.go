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

func TestMemorySubscriptionRegistry_FindMatchingConnections(t *testing.T) {
	tests := []struct {
		name  string
		event domain.Event
		setup func(*memory.MemorySubscriptionRegistry)
		want  []domain.ConnectionID
	}{
		{
			name:  "event with kind=1 matches specific connection",
			event: domain.Event{Kind: 1},
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				// 接続1: kinds=[1] のみを購読
				sub1 := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{{Kinds: []int{1}}},
				}
				msr.Register("conn-1", sub1)

				// 接続2: kinds=[2] のみを購読
				sub2 := domain.Subscription{
					ID:      "sub-2",
					Filters: []domain.Filter{{Kinds: []int{2}}},
				}
				msr.Register("conn-2", sub2)
			},
			want: []domain.ConnectionID{"conn-1"},
		},
		{
			name:  "multiple connections match",
			event: domain.Event{Kind: 1},
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				// 接続1: kinds=[1] を購読
				sub1 := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{{Kinds: []int{1}}},
				}
				msr.Register("conn-1", sub1)

				// 接続2: kinds=[1,2] を購読
				sub2 := domain.Subscription{
					ID:      "sub-2",
					Filters: []domain.Filter{{Kinds: []int{1, 2}}},
				}
				msr.Register("conn-2", sub2)
			},
			want: []domain.ConnectionID{"conn-1", "conn-2"},
		},
		{
			name:  "no filters match",
			event: domain.Event{Kind: 3},
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				sub := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{{Kinds: []int{1, 2}}},
				}
				msr.Register("conn-1", sub)
			},
			want: []domain.ConnectionID{},
		},
		{
			name:  "OR condition test with multiple filters",
			event: domain.Event{Kind: 2},
			setup: func(msr *memory.MemorySubscriptionRegistry) {
				// kinds=[1] OR kinds=[2] のフィルタ
				filter1 := domain.Filter{Kinds: []int{1}}
				filter2 := domain.Filter{Kinds: []int{2}}
				sub := domain.Subscription{
					ID:      "sub-1",
					Filters: []domain.Filter{filter1, filter2},
				}
				msr.Register("conn-1", sub)
			},
			want: []domain.ConnectionID{"conn-1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msr := memory.NewMemorySubscriptionRegistry()
			tt.setup(msr.(*memory.MemorySubscriptionRegistry))
			got := msr.FindMatchingConnections(tt.event)

			// スライスの順序を無視して比較
			if !equalConnectionIDs(got, tt.want) {
				t.Errorf("FindMatchingConnections() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ヘルパー関数：ConnectionIDスライスの比較（順序無視）
func equalConnectionIDs(a, b []domain.ConnectionID) bool {
	if len(a) != len(b) {
		return false
	}

	count := make(map[domain.ConnectionID]int)
	for _, id := range a {
		count[id]++
	}
	for _, id := range b {
		count[id]--
	}

	for _, c := range count {
		if c != 0 {
			return false
		}
	}
	return true
}
