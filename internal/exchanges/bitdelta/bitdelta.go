package bitdelta

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.bitdelta.com/open/api/v1/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitdelta" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	LastPrice   jsonutil.Float `json:"last_price"`
	QuoteVolume jsonutil.Float `json:"quote_volume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var data map[string]ticker
	if err := client.GetJSON(url, &data); err != nil {
		return err
	}
	ts := client.Now()
	for sym, t := range data {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: t.QuoteVolume.V()})
	}
	return nil
}
