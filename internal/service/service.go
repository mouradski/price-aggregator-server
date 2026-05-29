package service

import (
	"ftso-prices/internal/model"
	"ftso-prices/internal/server"
)

// TickerService is the single entry point for produced tickers: it drops
// tickers with a blank quote then broadcasts, mirroring TickerService.
type TickerService struct {
	server *server.TickerServer
}

func New(srv *server.TickerServer) *TickerService {
	return &TickerService{server: srv}
}

func (s *TickerService) Push(t model.Ticker) {
	if t.Quote == "" {
		return
	}
	s.server.Broadcast(t)
}
