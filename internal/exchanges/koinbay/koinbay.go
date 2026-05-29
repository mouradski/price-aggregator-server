package koinbay

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "koinbay" }

func (c *Client) URI(*client.Context) string { return "wss://ws.koinbay.com/kline-api/ws" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.Assets() {
		for _, quote := range ctx.Quotes() {
			send(fmt.Sprintf(`{"event":"sub","params":{"channel":"market_%s%s_ticker","cb_id":"1"}}`, base, quote))
		}
	}
}

// Pong answers Koinbay's ping by echoing it back as a pong.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "ping") {
		send(strings.ReplaceAll(message, "ping", "pong"))
		return true
	}
	return false
}

type tickerMsg struct {
	Channel string `json:"channel"`
	Tick    struct {
		Close    jsonutil.Float `json:"close"`
		Vol      jsonutil.Float `json:"vol"`
		AskPrice jsonutil.Float `json:"askPrice"`
		BidPrice jsonutil.Float `json:"bidPrice"`
	} `json:"tick"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "_ticker") {
		return nil
	}
	var m tickerMsg
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	parts := strings.Split(m.Channel, "_")
	if len(parts) < 2 {
		return nil
	}
	base, quote := symbol.GetPair(parts[1])
	volume := m.Tick.Vol.V() * m.Tick.Close.V()
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: m.Tick.Close.V(), H24Volume: volume},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (m.Tick.AskPrice.V() + m.Tick.BidPrice.V()) / 2, H24Volume: volume},
	}
}
