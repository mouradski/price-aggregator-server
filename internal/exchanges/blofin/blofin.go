package blofin

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Blofin (OKX-style) lists USDT/USDC markets. Note: the endpoint is blocked by
// Blofin's WAF (HTTP 403) from some hosting/VPN IPs but works from normal
// networks — confirmed locally with the same subscribe payload.
var wsQuotes = []string{"USDT", "USDC"}

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "blofin" }

func (c *Client) URI(*client.Context) string { return "wss://openapi.blofin.com/ws/public" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			send(fmt.Sprintf(`{"op":"subscribe","args":[{"channel":"tickers","instId":"%s-%s"}]}`, base, quote))
		}
	}
}

func (c *Client) PingInterval() time.Duration { return 30 * time.Second }

func (c *Client) Ping(send func(string)) { send("ping") }

type marketData struct {
	Data []struct {
		InstID   string         `json:"instId"`
		Last     jsonutil.Float `json:"last"`
		AskPrice jsonutil.Float `json:"askPrice"`
		BidPrice jsonutil.Float `json:"bidPrice"`
		Vol24h   jsonutil.Float `json:"vol24h"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"channel":"tickers"`) {
		return nil
	}
	var m marketData
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range m.Data {
		base, quote := symbol.GetPair(d.InstID)
		vol := d.Vol24h.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: d.Last.V(), H24Volume: vol},
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (d.AskPrice.V() + d.BidPrice.V()) / 2, H24Volume: vol})
	}
	return out
}
