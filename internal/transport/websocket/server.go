package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"nostar/internal/relay/domain"
	"nostar/internal/relay/usecase"

	"go.uber.org/zap"
)

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

	zap.S().Infow("starting websocket listener (HTTP stub)", "addr", s.addr)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		// 通常 Server Close で現れないエラーが発生した場合
		zap.S().Errorw("failed to close server", "err", err)
	}

	zap.S().Infow("close server", "addr", s.addr)
	return nil
}

// ServeHTTP is a placeholder: it pretends to receive one message via HTTP body,
// then routes it into RelayService. Replace with real WS handling later.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var wire WireMessage
	if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
		http.Error(w, "invalid message", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	switch wire.Type {
	case "EVENT":
		var evt domain.Event
		if err := json.Unmarshal(wire.Payload, &evt); err != nil {
			http.Error(w, "invalid event", http.StatusBadRequest)
			return
		}
		if err := s.relay.HandleEvent(ctx, usecase.EventMessage{Event: evt}); err != nil {
			http.Error(w, "failed to handle event", http.StatusInternalServerError)
			return
		}
	case "REQ":
		var sub domain.Subscription
		if err := json.Unmarshal(wire.Payload, &sub); err != nil {
			http.Error(w, "invalid subscription", http.StatusBadRequest)
			return
		}
		events, err := s.relay.HandleReq(ctx, usecase.ReqMessage{Subscription: sub})
		if err != nil {
			http.Error(w, "failed to handle req", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(events)
	case "CLOSE":
		var msg usecase.CloseMessage
		if err := json.Unmarshal(wire.Payload, &msg); err != nil {
			http.Error(w, "invalid close", http.StatusBadRequest)
			return
		}
		if err := s.relay.HandleClose(ctx, msg); err != nil {
			http.Error(w, "failed to close", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "unknown message type", http.StatusBadRequest)
		return
	}
}

type WireMessage struct {
	Type    string          `json:"type"`    // "EVENT", "REQ", "CLOSE"
	Payload json.RawMessage `json:"payload"` // raw JSON of each message type
}
