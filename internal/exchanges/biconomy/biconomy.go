package biconomy

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://www.biconomy.com/api/v1/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "biconomy" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Ticker []struct {
		Symbol string         `json:"symbol"`
		Last   jsonutil.Float `json:"last"`
		Vol    jsonutil.Float `json:"vol"`
		Buy    jsonutil.Float `json:"buy"`
		Sell   jsonutil.Float `json:"sell"`
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
		if t.Buy.V() != 0 && t.Sell.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (t.Sell.V() + t.Buy.V()) / 2, Timestamp: ts, H24Volume: vol})
		}
	}
	return nil
}
