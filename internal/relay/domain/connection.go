package domain

import (
	"sync"

	"github.com/google/uuid"
)

type ConnectionID string // サーバ生成、システム間ユニーク
type Connection interface {
	ID() ConnectionID
	WriteJSON(v interface{}) error
	Close() error
}

// 過度な抽象化を避けるために、抽象化せず struct にする
type ConnectionPool struct {
	mu    sync.RWMutex // 読み書きを別々でロックできる
	conns map[ConnectionID]Connection
}

func NewConnectionID() ConnectionID {
	return ConnectionID(uuid.New().String())
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		conns: make(map[ConnectionID]Connection),
	}
}

func (cp *ConnectionPool) Add(conn Connection) {
	cp.mu.Lock() // 書き込みロック
	defer cp.mu.Unlock()
	cp.conns[conn.ID()] = conn
}

func (cp *ConnectionPool) Remove(id ConnectionID) {
	cp.mu.Lock() // 書き込みロック
	defer cp.mu.Unlock()
	delete(cp.conns, id)
}

func (cp *ConnectionPool) Get(id ConnectionID) (Connection, bool) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	conn, exists := cp.conns[id]
	return conn, exists
}

// 新しいイベントを、興味を持っているクライアント（接続）全員に一斉配信する
func (cp *ConnectionPool) BroadcastTo(ids []ConnectionID, message interface{}) {
	cp.mu.RLock()         // 読み書きロック
	defer cp.mu.RUnlock() // 読み書きロック
	for _, id := range ids {
		if conn, exists := cp.conns[id]; exists {
			// エラーハンドリングは適宜
			go conn.WriteJSON(message)
		}
	}
}
