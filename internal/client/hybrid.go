package client

import (
	"sync/atomic"
	"time"

	"ftso-prices/internal/model"
)

// Health tracks when an exchange last produced a ticker over its websocket, so
// the REST poller can act purely as a startup seed and a fallback.
type Health struct {
	lastWS atomic.Int64 // epoch millis of the last WS-sourced ticker
}

func (h *Health) markWS() { h.lastWS.Store(time.Now().UnixMilli()) }

// wsFresh reports whether a WS ticker arrived within the last d.
func (h *Health) wsFresh(d time.Duration) bool {
	last := h.lastWS.Load()
	if last == 0 {
		return false
	}
	return time.Now().UnixMilli()-last < d.Milliseconds()
}

// HybridRunner runs a websocket source as the primary feed and a REST source as
// a startup seed + fallback. While the websocket is delivering tickers the REST
// poll is skipped; it resumes automatically if the websocket goes silent for
// longer than staleAfter.
type HybridRunner struct {
	ws   *WsRunner
	rest *RestRunner
	name string
}

// NewHybridRunner wires a WS and REST implementation of the same exchange to a
// shared health signal. push is the common downstream sink.
func NewHybridRunner(ws WsExchange, rest RestExchange, ctx *Context, push func(model.Ticker), staleAfter time.Duration) *HybridRunner {
	h := &Health{}

	wsPush := func(t model.Ticker) {
		h.markWS()
		push(t)
	}
	wsRunner := NewWsRunner(ws, ctx, wsPush)
	restRunner := NewRestRunner(rest, ctx, push)
	restRunner.gate = func() bool { return h.wsFresh(staleAfter) }

	return &HybridRunner{ws: wsRunner, rest: restRunner, name: ws.Name()}
}

func (r *HybridRunner) Name() string { return r.name }

// Run blocks until stop is closed, running both sources concurrently.
func (r *HybridRunner) Run(stop <-chan struct{}) {
	done := make(chan struct{}, 2)
	go func() { r.ws.Run(stop); done <- struct{}{} }()
	go func() { r.rest.Run(stop); done <- struct{}{} }()
	<-done
	<-done
}
