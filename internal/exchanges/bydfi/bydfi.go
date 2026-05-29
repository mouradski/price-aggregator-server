package bydfi

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://www.bydfi.com/b2b/rank/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bydfi" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

type response struct {
	Data map[string]struct {
		LastPrice   jsonutil.Float `json:"last_price"`
		QuoteVolume jsonutil.Float `json:"quote_volume"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for sym, d := range r.Data {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.LastPrice.V(), Timestamp: ts, H24Volume: d.QuoteVolume.V()})
	}
	return nil
}
