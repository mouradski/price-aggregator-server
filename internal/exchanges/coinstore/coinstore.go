package coinstore

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "coinstore" }

func (c *Client) URI(*client.Context) string { return "wss://ws.coinstore.com/s/ws" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.Assets() {
		for _, quote := range ctx.Quotes() {
			c.id++
			send(fmt.Sprintf(`{"op":"SUB","channel":["%s%s@ticker"],"id":%d}`, base, quote, c.id))
		}
	}
}

// Pong answers Coinstore's ping with a pong frame.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "ping") {
		send(fmt.Sprintf(`{"op":"pong","epochMillis":%d}`, client.Now()))
		return true
	}
	return false
}

type tickerData struct {
	Symbol string         `json:"symbol"`
	Close  jsonutil.Float `json:"close"`
	Volume jsonutil.Float `json:"volume"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "@ticker") {
		return nil
	}
	var d tickerData
	if err := json.Unmarshal([]byte(message), &d); err != nil || d.Symbol == "" {
		return nil
	}
	base, quote := symbol.GetPair(d.Symbol)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: d.Close.V(), H24Volume: d.Close.V() * d.Volume.V(),
	}}
}
