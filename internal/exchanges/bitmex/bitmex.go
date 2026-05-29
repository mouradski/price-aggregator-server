package bitmex

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitmex" }

func (c *Client) URI(*client.Context) string { return "wss://ws.bitmex.com/realtime" }

func (c *Client) Subscribe(_ *client.Context, send func(string)) {
	// Subscribe to the whole instrument table; non-matching symbols are filtered.
	send(`{"op":"subscribe","args":["instrument"]}`)
}

type wsMessage struct {
	Table string `json:"table"`
	Data  []struct {
		Symbol    string         `json:"symbol"`
		LastPrice jsonutil.Float `json:"lastPrice"`
		BidPrice  jsonutil.Float `json:"bidPrice"`
		AskPrice  jsonutil.Float `json:"askPrice"`
		Volume24h jsonutil.Float `json:"volume24h"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"table":"instrument"`) {
		return nil
	}
	var m wsMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range m.Data {
		if d.LastPrice.V() == 0 { // delta without a price update
			continue
		}
		// XBT is BitMEX's ticker for BTC.
		base, quote := symbol.GetPair(strings.ReplaceAll(d.Symbol, "XBT", "BTC"))
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Volume24h.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: d.LastPrice.V(), H24Volume: vol})
		if d.BidPrice.V() != 0 && d.AskPrice.V() != 0 {
			out = append(out, model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask",
				Base: base, Quote: quote, LastPrice: (d.BidPrice.V() + d.AskPrice.V()) / 2, H24Volume: vol})
		}
	}
	return out
}
