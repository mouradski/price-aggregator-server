package client

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	"ftso-prices/internal/model"
)

const (
	maxMessageBytes = 10 * 1024 * 1024
	reconnectDelay  = 20 * time.Second
)

// WsRunner drives a single websocket exchange: it connects, subscribes, reads
// messages (decompressing binary frames), enforces a no-message timeout and
// reconnects on failure. It is the Go equivalent of AbstractClientEndpoint's
// connection lifecycle.
type WsRunner struct {
	ex   WsExchange
	ctx  *Context
	push func(model.Ticker)

	writeMu sync.Mutex
}

func NewWsRunner(ex WsExchange, ctx *Context, push func(model.Ticker)) *WsRunner {
	return &WsRunner{ex: ex, ctx: ctx, push: push}
}

func (r *WsRunner) Name() string { return r.ex.Name() }

// Run blocks until stop is closed, reconnecting indefinitely between sessions.
func (r *WsRunner) Run(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}

		r.session(stop)

		select {
		case <-stop:
			return
		case <-time.After(reconnectDelay):
		}
	}
}

func (r *WsRunner) session(stop <-chan struct{}) {
	uri := r.ex.URI(r.ctx)
	log.Printf("Connecting to %s ...", r.ex.Name())

	var header http.Header
	if hp, ok := r.ex.(HeaderProvider); ok {
		header = hp.Headers()
	}

	conn, _, err := websocket.DefaultDialer.Dial(uri, header)
	if err != nil {
		log.Printf("Unable to connect to %s: %v", r.ex.Name(), err)
		return
	}
	defer conn.Close()
	conn.SetReadLimit(maxMessageBytes)
	log.Printf("Connected to %s", r.ex.Name())

	send := func(msg string) {
		r.writeMu.Lock()
		defer r.writeMu.Unlock()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			log.Printf("Error sending to %s: %v", r.ex.Name(), err)
		}
	}

	r.ex.Subscribe(r.ctx, send)

	var lastMsg atomic.Int64
	lastMsg.Store(time.Now().UnixMilli())
	done := make(chan struct{})
	defer close(done)

	go r.watchdog(conn, &lastMsg, done, stop)

	if p, ok := r.ex.(Pinger); ok {
		go r.pinger(p, send, done)
	}

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Connection to %s closed: %v", r.ex.Name(), err)
			return
		}
		lastMsg.Store(time.Now().UnixMilli())

		message := string(data)
		if mt == websocket.BinaryMessage {
			message = decompress(data)
		}
		r.process(message, send)
	}
}

// watchdog closes the connection if no message arrives within the timeout,
// mirroring checkMessageReceivedTimeout.
func (r *WsRunner) watchdog(conn *websocket.Conn, lastMsg *atomic.Int64, done, stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	timeoutMs := r.ctx.Timeout().Milliseconds()

	for {
		select {
		case <-done:
			return
		case <-stop:
			conn.Close()
			return
		case <-ticker.C:
			if time.Now().UnixMilli()-lastMsg.Load() > timeoutMs {
				log.Printf("No data received in a while from %s, reconnecting", r.ex.Name())
				conn.Close()
				return
			}
		}
	}
}

// pinger sends periodic application-level pings for exchanges that need them.
func (r *WsRunner) pinger(p Pinger, send func(string), done <-chan struct{}) {
	ticker := time.NewTicker(p.PingInterval())
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			p.Ping(send)
		}
	}
}

func (r *WsRunner) process(message string, send func(string)) {
	defer func() {
		// Mirror the broad catch in onMessage: a single malformed message must
		// never crash the reader loop.
		_ = recover()
	}()

	if p, ok := r.ex.(Ponger); ok {
		if p.Pong(message, send) {
			return
		}
	} else if defaultPong(message) {
		return
	}

	if d, ok := r.ex.(MetadataDecoder); ok {
		d.DecodeMetadata(message, send)
	}

	for _, t := range r.ex.MapTicker(r.ctx, message) {
		r.push(t)
	}
}

// FilterPush wraps a downstream push with the base-asset filter from
// AbstractClientEndpoint.pushTicker: only tickers whose (lowercased) base is in
// the configured asset set are forwarded.
func FilterPush(assetSet map[string]bool, down func(model.Ticker)) func(model.Ticker) {
	return func(t model.Ticker) {
		if t.Base == "" || !assetSet[strings.ToLower(t.Base)] {
			return
		}
		down(t)
	}
}
