package websocket

import (
	"nostar/internal/relay/domain"

	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	id   domain.ConnectionID
	conn *websocket.Conn
}

func (c *WebSocketConnection) ID() domain.ConnectionID       { return c.id }
func (c *WebSocketConnection) WriteJSON(v interface{}) error { return c.conn.WriteJSON(v) }
func (c *WebSocketConnection) Close() error                  { return c.conn.Close() }
