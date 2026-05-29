package huobi

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "huobi" }

func (c *Client) URI(*client.Context) string { return "wss://api.huobi.pro/ws" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, asset := range ctx.Assets() {
		for _, quote := range ctx.Quotes() {
			send(fmt.Sprintf(`{"sub":"market.%s%s.ticker","id":"%d"}`, asset, quote, client.Now()))
		}
	}
}

// Pong answers Huobi's JSON heartbeat by echoing it back as a pong.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "ping") {
		send(strings.ReplaceAll(message, "ping", "pong"))
		return true
	}
	return false
}

type response struct {
	Ch   string `json:"ch"`
	Tick struct {
		LastPrice float64 `json:"lastPrice"`
		Ask       float64 `json:"ask"`
		Bid       float64 `json:"bid"`
		Vol       float64 `json:"vol"`
	} `json:"tick"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker") || !strings.Contains(message, "open") {
		return nil
	}
	var r response
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}

	parts := strings.Split(r.Ch, ".")
	if len(parts) < 2 {
		return nil
	}
	base, quote := symbol.GetPair(strings.ToUpper(parts[1]))
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: r.Tick.LastPrice, H24Volume: r.Tick.Vol},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (r.Tick.Ask + r.Tick.Bid) / 2, H24Volume: r.Tick.Vol},
	}
}
