package domain

type SubscriptionRegistry interface {
	Register(connID ConnectionID, sub Subscription) error // この connID で subscription を追加
	Unregister(connID ConnectionID, subID string) error   // この connID の subscription を削除
	UnregisterAll(connID ConnectionID) error              // この connID の subscription を全削除
	FindMatchingConnections(event Event) []ConnectionID   // このイベントに興味を持っているクライアント（接続）はどれかを特定
}
