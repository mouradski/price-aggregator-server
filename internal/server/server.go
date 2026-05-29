package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"ftso-prices/internal/model"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}

// TickerServer broadcasts tickers as JSON to every connected websocket client,
// mirroring TickerServer/WsServer in the Java implementation.
type TickerServer struct {
	mu       sync.RWMutex
	sessions map[*session]struct{}
}

type session struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func New() *TickerServer {
	return &TickerServer{sessions: make(map[*session]struct{})}
}

// Handler serves the /ticker websocket endpoint.
func (s *TickerServer) Handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	sess := &session{conn: conn}

	s.mu.Lock()
	s.sessions[sess] = struct{}{}
	s.mu.Unlock()
	log.Printf("Client connected to ticker channel")

	// Drain client messages until it disconnects, then clean up.
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.sessions, sess)
			s.mu.Unlock()
			conn.Close()
			log.Printf("Client disconnected from ticker channel")
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

// Broadcast serialises the ticker and sends it to all connected clients.
func (s *TickerServer) Broadcast(t model.Ticker) {
	payload, err := json.Marshal(t)
	if err != nil {
		return
	}

	s.mu.RLock()
	targets := make([]*session, 0, len(s.sessions))
	for sess := range s.sessions {
		targets = append(targets, sess)
	}
	s.mu.RUnlock()

	for _, sess := range targets {
		sess.mu.Lock()
		err := sess.conn.WriteMessage(websocket.TextMessage, payload)
		sess.mu.Unlock()
		if err != nil {
			s.mu.Lock()
			delete(s.sessions, sess)
			s.mu.Unlock()
			sess.conn.Close()
		}
	}
}
