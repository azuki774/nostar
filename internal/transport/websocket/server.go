package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

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
	addr  string // 例: 127.0.0.1:9999
	relay *usecase.RelayService
}

func NewServer(addr string, relay *usecase.RelayService) *Server {
	return &Server{
		addr:  addr,
		relay: relay,
	}
}

// Run starts an HTTP server that would upgrade connections to WebSocket.
// The actual WS upgrade/loop is TODO; this skeleton shows dependency wiring.
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
	// ここで HTTP → WebSocket にアップグレード
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		zap.S().Errorw("upgrade error", zap.Error(err))
		return
	}
	defer c.Close()
	zap.S().Debugw("websocket upgraded", "remote_addr", r.RemoteAddr)

	for {
		ctx := r.Context()
		_, data, err := c.ReadMessage()
		if err != nil {
			zap.S().Infow("websocket read closed", "err", err)
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

		case "CLOSE":
			zap.S().Debugw("received CLOSE")
			// wire.SubscriptionID を usecase.CloseMessage に渡す
		}
	}

}
