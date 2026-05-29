package hashkey

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

func (c *Client) Name() string { return "hashkey" }

func (c *Client) URI(*client.Context) string { return "wss://stream-pro.hashkey.com/quote/ws/v1" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			c.id++
			send(fmt.Sprintf(`{"symbol":"%s%s","topic":"realtimes","event":"sub","params":{"binary":"False"},"id":%d}`,
				base, quote, c.id))
		}
	}
}

func (c *Client) PingInterval() time.Duration { return 10 * time.Second }

func (c *Client) Ping(send func(string)) {
	send(fmt.Sprintf(`{"ping":%d}`, client.Now()))
}

type marketData struct {
	Symbol string `json:"symbol"`
	Data   []struct {
		C  jsonutil.Float `json:"c"`
		Qv jsonutil.Float `json:"qv"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "realtimes") {
		return nil
	}
	var m marketData
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(m.Symbol)
	ts := client.Now()
	var out []model.Ticker
	for _, d := range m.Data {
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.C.V(), H24Volume: d.Qv.V(),
		})
	}
	return out
}
