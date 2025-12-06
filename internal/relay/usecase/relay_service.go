package usecase

import (
	"context"
	"fmt"

	"nostar/internal/infrastructure/memory"
	"nostar/internal/relay"
	"nostar/internal/relay/domain"

	"go.uber.org/zap"
)

// RelayService bundles the Nostr relay business use cases.
// It does not know about transports (WebSocket/HTTP); those call into this type.
type RelayService struct {
	store    relay.EventStore
	registry domain.SubscriptionRegistry
	connPool domain.ConnectionPool
}

func NewRelayService(store relay.EventStore) *RelayService {
	return &RelayService{
		store:    store,
		registry: memory.NewMemorySubscriptionRegistry(),
	}
}

// HandleEvent processes an EVENT message: validation, persistence, and fanout.
func (s *RelayService) HandleEvent(ctx context.Context, msg EventMessage) error {
	zap.S().Debugw("HandleEvent called", "event_id", msg.Event.ID, "kind", msg.Event.Kind)
	// Validate event fields
	if err := msg.Event.Validate(); err != nil {
		return err
	}

	// Verify signature and ID
	valid, err := msg.Event.CheckSignature()
	if err != nil {
		return err
	}
	if !valid {
		return err
	}

	// Save to store
	if err := s.store.Save(ctx, msg.Event); err != nil {
		return err
	}

	// 関心のある subscribers （connectionID含む）を取得
	subs := s.registry.FindMatchingSubscriptions(msg.Event)
	// 新しいイベントをブロードキャストする
	if err := s.BroadcastToSubscribers(ctx, msg.Event, subs); err != nil {
		zap.S().Errorw("broadcast error", "err", err.Error())
		return err
	}
	return nil
}

// HandleReq processes a REQ: query stored events and start live subscription if available.
func (s *RelayService) HandleReq(ctx context.Context, msg ReqMessage) ([]domain.Event, error) {
	events, err := s.store.Query(ctx, msg.Subscription)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// HandleClose processes CLOSE; any subscription cleanup would happen here.
func (s *RelayService) HandleClose(ctx context.Context, msg CloseMessage) error {
	return nil
}

// 既に subscriptionID が存在する場合は上書きする
func (s *RelayService) RegisterSubscription(ctx context.Context, msg ReqMessage) error {
	// まず既存のサブスクリプションを解除（存在しなくてもエラーにならない）
	_ = s.registry.Unregister(msg.ConnectionID, msg.Subscription.ID)

	return s.registry.Register(msg.ConnectionID, msg.Subscription)
}

func (s *RelayService) UnregisterSubscription(ctx context.Context, msg CloseMessage) error {
	return s.registry.Unregister(msg.ConnectionID, msg.SubscriptionID)
}

func (s *RelayService) UnregisterAllSubscriptions(ctx context.Context, connID domain.ConnectionID) error {
	return s.registry.UnregisterAll(connID)
}

func (s *RelayService) BroadcastToSubscribers(ctx context.Context, evt domain.Event, subs []domain.SubscriptionMatch) error {
	zap.S().Infow("BroadcastToSubscribers called", "subscriber_count", len(subs))
	for _, sub := range subs {
		conn, exists := s.connPool.Get(sub.ConnectionID)
		if !exists {
			// 接続が存在しない場合はスキップ（切断済みの場合）
			zap.S().Infow("connection is lost", "connID", sub.ConnectionID)
			continue
		}

		eventMsg := []any{"EVENT", sub.SubscriptionID, evt}
		zap.S().Debugw("sending event", "connID", sub.ConnectionID)
		if err := conn.WriteJSON(eventMsg); err != nil {
			// 送信失敗時はエラーを返す（最初のエラーを返す）
			return fmt.Errorf("failed to send event to connection %s: %w", sub.ConnectionID, err)
		}
	}
	return nil
}
