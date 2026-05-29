package orangex

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

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "orangex" }

func (c *Client) URI(*client.Context) string { return "wss://api.orangex.com/ws/api/v1" }

// OrangeX lists USDT/USDC markets and accepts only ONE channel per subscribe
// call. It also rate-limits inbound messages and drops the connection (1006) if
// flooded, so the per-channel subscribes are paced.
var wsQuotes = []string{"USDT", "USDC"}

const subscribePacing = 30 * time.Millisecond

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			c.id++
			send(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"/public/subscribe","params":{"channels":["ticker.%s-%s.raw"]}}`,
				c.id, base, quote))
			time.Sleep(subscribePacing)
		}
	}
}

func (c *Client) PingInterval() time.Duration { return 10 * time.Second }

func (c *Client) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"/public/ping","params":{}}`, c.id))
}

type ticker struct {
	Params struct {
		Data struct {
			LastPrice      jsonutil.Float `json:"last_price"`
			InstrumentName string         `json:"instrument_name"`
			BestBidPrice   jsonutil.Float `json:"best_bid_price"`
			BestAskPrice   jsonutil.Float `json:"best_ask_price"`
			Stats          struct {
				Volume jsonutil.Float `json:"volume"`
			} `json:"stats"`
		} `json:"data"`
	} `json:"params"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "last_price") {
		return nil
	}
	var t ticker
	if err := json.Unmarshal([]byte(message), &t); err != nil {
		return nil
	}
	d := t.Params.Data
	base, quote := symbol.GetPair(d.InstrumentName)
	volume := d.Stats.Volume.V() * d.LastPrice.V()
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.LastPrice.V(), H24Volume: volume},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.BestAskPrice.V() + d.BestBidPrice.V()) / 2, H24Volume: volume},
	}
}
