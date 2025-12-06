package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"nostar/internal/config"
	"nostar/internal/relay/domain"
	"nostar/internal/relay/usecase"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO
	},
}

// Server is an inbound adapter that accepts WebSocket connections
// and forwards parsed messages to RelayService.
type Server struct {
	addr           string // 例: 127.0.0.1:9999
	relay          *usecase.RelayService
	connectionPool *domain.ConnectionPool
	relayInfo      *config.RelayInfoConfig
}

func NewServer(addr string, relay *usecase.RelayService, connPool *domain.ConnectionPool, relayInfo *config.RelayInfoConfig) *Server {
	return &Server{
		addr:           addr,
		relay:          relay,
		connectionPool: connPool,
		relayInfo:      relayInfo,
	}
}

// Run starts an HTTP server that would upgrade connections to WebSocket.
func (s *Server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.Handle("/", s) // Server implements http.Handler below.

	srv := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	// Note: proper graceful shutdown should hook into ctx.Done().
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(ctx)
	}()

	zap.S().Infow("starting websocket listener", "addr", s.addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// 通常 Server Close で現れないエラーが発生した場合
		zap.S().Errorw("failed to close server", "err", err)
	}

	zap.S().Infow("close server", "addr", s.addr)
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// NIP-11: Relay Information Document
	if r.Method == "GET" && (r.URL.Path == "/.well-known/nostr.json") {
		s.handleRelayInfo(w, r)
		return
	}

	// ここで HTTP → WebSocket にアップグレード
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().Errorw("upgrade error", zap.Error(err))
		return
	}
	defer c.Close()

	connID := domain.NewConnectionID()
	zap.S().Debugw("websocket upgraded", "remote_addr", r.RemoteAddr)

	// WebSocketConnection を作成
	wsConn := &WebSocketConnection{
		id:   connID,
		conn: c,
	}

	// ConnectionPool に追加
	s.connectionPool.Add(wsConn)
	defer func() {
		// 接続切断時にすべてのサブスクリプションを解除
		if err := s.relay.UnregisterAllSubscriptions(context.Background(), connID); err != nil {
			zap.S().Errorw("failed to unregister all subscriptions", "connID", connID, "error", err)
		}
		s.connectionPool.Remove(connID)
	}()

	zap.S().Debugw("added to connection pool", "id", connID, "num", s.connectionPool.GetSize())

	for {
		ctx := r.Context()
		_, data, err := c.ReadMessage()
		if err != nil {
			var ce *websocket.CloseError
			if errors.As(err, &ce) && ce.Code == websocket.CloseAbnormalClosure {
				// クライアント or ネットワーク都合の異常終了として info 扱い
				zap.S().Infow("websocket closed by client", "code", ce.Code, "text", ce.Text)
			} else {
				// それ以外はエラーとして扱う
				zap.S().Errorw("websocket read error", "err", err)
			}
			return
		}

		var wire WireMessage
		if err := json.Unmarshal(data, &wire); err != nil {
			zap.S().Debugw("unknown data", zap.String("data", string(data)), zap.Error(err))
			if err := c.WriteJSON([]string{"NOTICE", "invalid JSON: cannot parse message"}); err != nil {
				zap.S().Errorw("write notice failed", zap.Error(err))
				return
			}
		}

		switch wire.Type {
		case "EVENT":
			var evt domain.Event
			if err := json.Unmarshal(wire.Event, &evt); err != nil {
				c.WriteJSON([]string{"NOTICE", "invalid JSON: cannot parse message"})
				continue
			}
			zap.S().Debugw("received EVENT", zap.Int("kind", evt.Kind))

			if err := s.relay.HandleEvent(ctx, usecase.EventMessage{Event: evt}); err != nil {
				zap.S().Errorw("handle EVENT failed", zap.Error(err))
				if writeErr := c.WriteJSON([]any{"OK", evt.ID, false, "internal error"}); writeErr != nil {
					// クライアントに EVENT 登録に失敗したことを通知
					zap.S().Errorw("write EVENT OK failed", zap.Error(writeErr))
					return
				}
				continue
			}

			if err := c.WriteJSON([]any{"OK", evt.ID, true, ""}); err != nil {
				zap.S().Errorw("write EVENT OK failed", zap.Error(err))
				return
			}

		case "REQ":
			zap.S().Debugw("received REQ")
			// wire.SubscriptionID, wire.Filters を使って Subscription を組み立てる

			// フィルタ部分だけ軽くパースする（1つ目だけを採用）
			// TODO: 複数フィルターに対応
			var f struct {
				Authors []string `json:"authors"`
				Kinds   []int    `json:"kinds"`
				Since   *int64   `json:"since"`
				Until   *int64   `json:"until"`
				Limit   *int     `json:"limit"`
			}

			if len(wire.Filters) > 0 {
				// TODO: 複数フィルターに対応
				if err := json.Unmarshal(wire.Filters[0], &f); err != nil {
					zap.S().Debugw("invalid REQ filter", "data", string(wire.Filters[0]), zap.Error(err))
					if err := c.WriteJSON([]string{"NOTICE", "invalid REQ filter"}); err != nil {
						zap.S().Errorw("write notice failed", zap.Error(err))
						return
					}
					continue // コネクションは継続
				}
			}

			filters, err := domain.NewFiltersFromRaw(wire.Filters)
			if err != nil {
				if err := c.WriteJSON([]string{"NOTICE", "invalid REQ filter"}); err != nil {
					zap.S().Errorw("failed to parse filters", zap.Error(err))
					return
				}
				continue // コネクションは継続
			}

			sub := domain.Subscription{
				ID:      wire.SubscriptionID,
				Filters: filters,
			}

			var events []domain.Event
			if events, err = s.relay.HandleReq(ctx, usecase.ReqMessage{Subscription: sub}); err != nil {
				zap.S().Errorw("handle REQ failed", zap.Error(err))
				if err := c.WriteJSON([]string{"NOTICE", "internal error on REQ"}); err != nil {
					zap.S().Errorw("write notice failed", zap.Error(err))
					return
				}
				continue
			}

			// 取得したイベントをクライアントに送信
			for _, evt := range events {
				if err := c.WriteJSON([]any{"EVENT", wire.SubscriptionID, evt}); err != nil {
					zap.S().Errorw("write event failed", zap.Error(err))
					return
				}
			}

			// この REQ に対する過去イベントの送信終了
			if err := c.WriteJSON([]any{"EOSE", wire.SubscriptionID}); err != nil {
				zap.S().Errorw("write EOSE failed", zap.Error(err))
				return
			}

			// この時点からライブ配信開始
			reqMsg := usecase.ReqMessage{
				Subscription: sub,
				ConnectionID: connID,
			}

			// SubscriptionRegistry に登録
			if err := s.relay.RegisterSubscription(ctx, reqMsg); err != nil {
				zap.S().Errorw("failed to error subscription", zap.Error(err))
				if err := c.WriteJSON([]string{"NOTICE", "internal error on REQ (subscription)"}); err != nil {
					zap.S().Errorw("write subscription failed", zap.Error(err))
				}
				return
			}
			zap.S().Infow("register subscription", "connID", connID, "subscriptionID", sub.ID)

		case "CLOSE":
			zap.S().Debugw("received CLOSE", "connID", connID, "subscriberID", wire.SubscriptionID)

			if err := s.relay.HandleClose(ctx, usecase.CloseMessage{SubscriptionID: wire.SubscriptionID}); err != nil {
				zap.S().Errorw("handle CLOSE failed", zap.Error(err))
				// CLOSE は ack 不要なので、エラーでもクライアントには特に何も返さない
			}

			closeMsg := usecase.CloseMessage{
				ConnectionID:   connID,
				SubscriptionID: wire.SubscriptionID,
			}

			if err := s.relay.UnregisterSubscription(ctx, closeMsg); err != nil {
				zap.S().Errorw("delete subscription failed", zap.Error(err))
				// ユーザに通知しなくていい
			}
		}
	}

}

// handleRelayInfo handles NIP-11 Relay Information Document requests
func (s *Server) handleRelayInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	relayInfo := map[string]interface{}{
		"name":        s.relayInfo.Name,
		"description": s.relayInfo.Description,
		"software":    s.relayInfo.Software,
		"version":     s.relayInfo.Version,
		// "limitation": map[string]interface{}{
		// 	"max_message_length": s.relayInfo.Limitations.MaxMessageLength,
		// 	"max_subscriptions":  s.relayInfo.Limitations.MaxSubscriptions,
		// 	"max_filters":        s.relayInfo.Limitations.MaxFilters,
		// 	"max_limit":          s.relayInfo.Limitations.MaxLimit,
		// 	"max_subid_length":   s.relayInfo.Limitations.MaxSubIDLength,
		// 	"min_pow_difficulty": s.relayInfo.Limitations.MinPowDifficulty,
		// 	"auth_required":      s.relayInfo.Limitations.AuthRequired,
		// 	"payment_required":   s.relayInfo.Limitations.PaymentRequired,
		// },
	}

	// Optional fields
	if s.relayInfo.Pubkey != "" {
		relayInfo["pubkey"] = s.relayInfo.Pubkey
	}
	if s.relayInfo.Contact != "" {
		relayInfo["contact"] = s.relayInfo.Contact
	}
	if len(s.relayInfo.SupportedNIPs) > 0 {
		relayInfo["supported_nips"] = s.relayInfo.SupportedNIPs
	}
	if len(s.relayInfo.RelayCountries) > 0 {
		relayInfo["relay_countries"] = s.relayInfo.RelayCountries
	}
	if len(s.relayInfo.LanguageTags) > 0 {
		relayInfo["language_tags"] = s.relayInfo.LanguageTags
	}
	// if len(s.relayInfo.Tags.List) > 0 {
	// 	relayInfo["tags"] = s.relayInfo.Tags.List
	// }
	if s.relayInfo.PostingPolicy != "" {
		relayInfo["posting_policy"] = s.relayInfo.PostingPolicy
	}

	if err := json.NewEncoder(w).Encode(relayInfo); err != nil {
		zap.S().Errorw("failed to encode relay info", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
