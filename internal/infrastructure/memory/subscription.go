package memory

import (
	"nostar/internal/relay/domain"
	"sync"
)

type MemorySubscriptionRegistry struct {
	mu   sync.RWMutex
	subs map[domain.ConnectionID][]domain.Subscription
}

// Register: 指定された接続IDにサブスクリプションを追加
func (msr *MemorySubscriptionRegistry) Register(connID domain.ConnectionID, sub domain.Subscription) error {
	msr.mu.Lock()
	defer msr.mu.Unlock()

	// 既存のサブスクリプションに追加
	msr.subs[connID] = append(msr.subs[connID], sub)
	return nil
}

// Unregister: 指定された接続IDから特定のサブスクリプションIDを削除
func (msr *MemorySubscriptionRegistry) Unregister(connID domain.ConnectionID, subID string) error {
	msr.mu.Lock() // 書き込みロック
	defer msr.mu.Unlock()

	subscriptions, exists := msr.subs[connID]
	if !exists {
		return nil // 既に存在しない場合は何もしない
	}

	// 指定された subID のサブスクリプションを削除
	for i, sub := range subscriptions {
		if sub.ID == subID {
			// スライスから削除
			msr.subs[connID] = append(subscriptions[:i], subscriptions[i+1:]...)
			break
		}
	}
	return nil
}

// UnregisterAll: 指定された接続IDの全てのサブスクリプションを削除
func (msr *MemorySubscriptionRegistry) UnregisterAll(connID domain.ConnectionID) error {
	msr.mu.Lock() // 書き込みロック
	defer msr.mu.Unlock()

	delete(msr.subs, connID)
	return nil
}

// FindMatchingConnections: 指定されたイベントにマッチするサブスクリプションを持つ全ての接続IDを返す
func (msr *MemorySubscriptionRegistry) FindMatchingConnections(event domain.Event) []domain.ConnectionID {
	msr.mu.RLock()         // 読み取りロック（書き込みOK）
	defer msr.mu.RUnlock() // 読み取りロック解除

	var matchingConnIDs []domain.ConnectionID

	// 全ての接続をチェック
	for connID, subscriptions := range msr.subs {
		// この接続のいずれかのサブスクリプションがイベントにマッチするか？
		for _, sub := range subscriptions {
			if sub.Matches(event) {
				matchingConnIDs = append(matchingConnIDs, connID)
				break // この接続はマッチしたので、次の接続へ
			}
		}
	}

	return matchingConnIDs
}

func NewMemorySubscriptionRegistry() domain.SubscriptionRegistry {
	return &MemorySubscriptionRegistry{
		subs: make(map[domain.ConnectionID][]domain.Subscription),
	}
}
