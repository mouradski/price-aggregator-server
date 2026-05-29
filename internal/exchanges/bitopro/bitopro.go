package bitopro

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.bitopro.com/v3/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitopro" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	Data []struct {
		Pair       string         `json:"pair"`
		LastPrice  jsonutil.Float `json:"lastPrice"`
		Volume24hr jsonutil.Float `json:"volume24hr"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data {
		base, quote := symbol.GetPair(t.Pair)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: t.Volume24hr.V() * t.LastPrice.V()})
	}
	return nil
}
