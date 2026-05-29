package hotcoin

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.hotcoinfin.com/v1/market/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "hotcoin" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Ticker []struct {
		Symbol string         `json:"symbol"`
		Last   jsonutil.Float `json:"last"`
		Sell   jsonutil.Float `json:"sell"`
		Buy    jsonutil.Float `json:"buy"`
		Vol    jsonutil.Float `json:"vol"`
	} `json:"ticker"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Ticker {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Vol.V() * t.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Buy.V() + t.Sell.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
