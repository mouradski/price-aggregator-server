package binanceus

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const wsBase = "wss://stream.binance.us:9443/stream?streams="

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "binanceus" }

func (c *Client) URI(ctx *client.Context) string {
	var streams []string
	for _, asset := range ctx.Assets() {
		for _, quote := range ctx.Quotes() {
			if quote == asset {
				continue
			}
			streams = append(streams, asset+quote+"@ticker")
		}
	}
	return wsBase + strings.Join(streams, "/")
}

func (c *Client) Subscribe(*client.Context, func(string)) {}

// Binance payloads carry both lowercase (c,b,a,q) and uppercase (C,B,A,Q) keys;
// the decoys absorb the uppercase ones (Go's JSON matching is case-insensitive).
type event struct {
	Data struct {
		S string         `json:"s"`
		C jsonutil.Float `json:"c"`
		B jsonutil.Float `json:"b"`
		A jsonutil.Float `json:"a"`
		Q jsonutil.Float `json:"q"`

		CloseTime jsonutil.Float `json:"C"`
		LastQty   jsonutil.Float `json:"Q"`
		BidQty    jsonutil.Float `json:"B"`
		AskQty    jsonutil.Float `json:"A"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker") {
		return nil
	}
	var e event
	if err := json.Unmarshal([]byte(message), &e); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(e.Data.S)
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: e.Data.C.V(), H24Volume: e.Data.Q.V()},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (e.Data.A.V() + e.Data.B.V()) / 2, H24Volume: e.Data.Q.V()},
	}
}
