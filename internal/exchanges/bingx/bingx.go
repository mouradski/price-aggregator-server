package bingx

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

func (c *Client) Name() string { return "bingx" }

func (c *Client) URI(*client.Context) string { return "wss://open-api-ws.bingx.com/market" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			c.id++
			send(fmt.Sprintf(`{"id":"%d","reqType":"sub","dataType":"%s-%s@ticker"}`, c.id, base, quote))
		}
	}
}

// Pong answers BingX's "Ping" with a literal "Pong".
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "Ping") {
		send("Pong")
		return true
	}
	return false
}

type response struct {
	Data struct {
		LastPrice jsonutil.Float `json:"c"`
		Symbol    string         `json:"s"`
		Volume    jsonutil.Float `json:"q"`
		Ask       jsonutil.Float `json:"A"`
		Bid       jsonutil.Float `json:"B"`

		// Decoys absorbing case-variant keys (see binance for the rationale).
		CloseTime jsonutil.Float `json:"C"`
		SymbolUp  string         `json:"S"`
		LastQty   jsonutil.Float `json:"Q"`
		AskPx     jsonutil.Float `json:"a"`
		BidPx     jsonutil.Float `json:"b"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "@ticker") {
		return nil
	}
	var r response
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(r.Data.Symbol)
	vol := r.Data.Volume.V()
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: r.Data.LastPrice.V(), H24Volume: vol},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (r.Data.Ask.V() + r.Data.Bid.V()) / 2, H24Volume: vol},
	}
}
