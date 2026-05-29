package ascendex

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Ascendex only quotes against USDT and USDC.
var stablecoinQuotes = []string{"USDT", "USDC"}

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "ascendex" }

func (c *Client) URI(*client.Context) string { return "wss://ascendex.com/0/api/pro/v1/stream" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range stablecoinQuotes {
			c.id++
			send(fmt.Sprintf(`{"op":"sub","id":"%d","ch":"summary:%s/%s"}`, c.id, base, quote))
		}
	}
}

// Pong answers Ascendex's ping.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, `"ping"`) {
		send(`{"op":"pong"}`)
		return true
	}
	return false
}

type summary struct {
	S    string `json:"s"`
	Data struct {
		C jsonutil.Float `json:"c"`
		V jsonutil.Float `json:"v"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "summary") {
		return nil
	}
	var s summary
	if err := json.Unmarshal([]byte(message), &s); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(s.S)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: s.Data.C.V(), H24Volume: s.Data.C.V() * s.Data.V.V(),
	}}
}
