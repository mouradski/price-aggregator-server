package bluebit

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://openapi.bluebit.io/open/api/get_allticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bluebit" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data struct {
		Ticker []struct {
			Symbol string         `json:"symbol"`
			Last   jsonutil.Float `json:"last"`
			Vol    jsonutil.Float `json:"vol"` // base volume
		} `json:"ticker"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data.Ticker {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: t.Vol.V() * t.Last.V()})
	}
	return nil
}
