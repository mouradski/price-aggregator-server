package client

import (
	"net/http"
	"strings"
	"time"

	"ftso-prices/internal/config"
	"ftso-prices/internal/model"
)

// Context carries the shared configuration that exchange implementations need
// to build subscriptions and filter pairs. It mirrors the helper methods on
// AbstractClientEndpoint (getAssets, getAllQuotes, getTimeout).
type Context struct {
	cfg      *config.Config
	quoteSet map[string]bool
}

func NewContext(cfg *config.Config) *Context {
	quoteSet := make(map[string]bool, len(cfg.Quotes))
	for _, q := range cfg.Quotes {
		quoteSet[q] = true
	}
	return &Context{cfg: cfg, quoteSet: quoteSet}
}

// IsAsset reports whether base (any case) is a configured asset.
func (c *Context) IsAsset(base string) bool { return c.cfg.AssetSet[strings.ToLower(base)] }

// IsQuote reports whether quote (any case) is a configured quote currency.
func (c *Context) IsQuote(quote string) bool { return c.quoteSet[strings.ToLower(quote)] }

// Assets returns the configured assets in lowercase.
func (c *Context) Assets() []string { return c.cfg.Assets }

// AssetsUpper returns the configured assets in uppercase.
func (c *Context) AssetsUpper() []string { return c.cfg.AssetsUpper() }

// Quotes returns the quote currencies in lowercase.
func (c *Context) Quotes() []string { return c.cfg.Quotes }

// QuotesUpper returns the quote currencies in uppercase.
func (c *Context) QuotesUpper() []string { return c.cfg.QuotesUpper() }

// Timeout is the no-message reconnect threshold.
func (c *Context) Timeout() time.Duration { return c.cfg.Timeout }

// Now returns the current epoch milliseconds, matching currentTimestamp().
func Now() int64 { return time.Now().UnixMilli() }

// WsExchange is implemented by websocket-driven exchanges.
type WsExchange interface {
	Name() string
	URI(ctx *Context) string
	Subscribe(ctx *Context, send func(string))
	MapTicker(ctx *Context, message string) []model.Ticker
}

// Ponger is an optional interface for exchanges with custom heartbeat handling.
// It returns true when the message was a ping/pong and must not be mapped.
type Ponger interface {
	Pong(message string, send func(string)) bool
}

// Pinger is an optional interface for exchanges that require the client to send
// periodic application-level pings to keep the connection alive.
type Pinger interface {
	PingInterval() time.Duration
	Ping(send func(string))
}

// HeaderProvider is an optional interface for exchanges whose websocket
// handshake needs custom HTTP headers (e.g. Origin / User-Agent).
type HeaderProvider interface {
	Headers() http.Header
}

// MetadataDecoder is an optional interface for exchanges that must extract state
// from non-ticker messages (e.g. channel/subscription id mappings) before
// mapping tickers. It mirrors AbstractClientEndpoint.decodeMetadata and is only
// ever called from the single per-connection read goroutine.
type MetadataDecoder interface {
	DecodeMetadata(message string, send func(string))
}

// RestExchange is implemented by REST-polling exchanges.
type RestExchange interface {
	Name() string
	Interval() time.Duration
	Poll(ctx *Context, push func(model.Ticker)) error
}

// defaultPong mirrors AbstractClientEndpoint.pong: short messages mentioning
// ping/pong are treated as heartbeats and ignored.
func defaultPong(message string) bool {
	lc := strings.ToLower(message)
	return len(lc) < 100 && (strings.Contains(lc, "ping") || strings.Contains(lc, "pong"))
}
