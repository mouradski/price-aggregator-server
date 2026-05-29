package binance

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const wsBase = "wss://stream.binance.com:9443/stream?streams="

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "binance" }

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

// Go's encoding/json matches keys case-insensitively, and Binance's payload
// has both lowercase (c, b, a, q) and uppercase (C, B, A, Q) keys. The decoy
// fields capture the uppercase keys so they cannot overwrite the lowercase ones.
type event struct {
	Data struct {
		S string         `json:"s"` // symbol
		C jsonutil.Float `json:"c"` // last price
		B jsonutil.Float `json:"b"` // best bid price
		A jsonutil.Float `json:"a"` // best ask price
		Q jsonutil.Float `json:"q"` // total traded quote volume

		// Decoys: capture the uppercase keys (must be exported for the decoder
		// to route them here instead of onto the lowercase fields).
		CloseTime int64          `json:"C"`
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
