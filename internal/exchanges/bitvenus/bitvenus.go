package bitvenus

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

// Bitvenus only quotes against USDT and USDC.
var stablecoinQuotes = []string{"USDT", "USDC"}

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitvenus" }

func (c *Client) URI(*client.Context) string { return "wss://wsapi.bitvenus.me/openapi/quote/ws/v1" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range stablecoinQuotes {
			send(fmt.Sprintf(`{"symbol":"%s%s","topic":"realtimes","event":"sub","params":{"binary":"false"}}`, base, quote))
		}
	}
}

func (c *Client) PingInterval() time.Duration { return 290 * time.Second }

func (c *Client) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"ping":%d}`, c.id))
}

// Pong ignores Bitvenus pong frames.
func (c *Client) Pong(message string, _ func(string)) bool {
	return strings.Contains(message, "pong")
}

type tickerResponse struct {
	Symbol string `json:"symbol"`
	Data   []struct {
		C jsonutil.Float `json:"c"`
		V jsonutil.Float `json:"v"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "realtimes") {
		return nil
	}
	var r tickerResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(r.Symbol)
	ts := client.Now()
	var out []model.Ticker
	for _, d := range r.Data {
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.C.V(), H24Volume: d.V.V() * d.C.V(),
		})
	}
	return out
}
