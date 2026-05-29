package coinex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.coinex.com/v1/market/ticker/all"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "coinex" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	Data struct {
		Ticker map[string]struct {
			Vol  jsonutil.Float `json:"vol"`
			Last jsonutil.Float `json:"last"`
			Buy  jsonutil.Float `json:"buy"`
			Sell jsonutil.Float `json:"sell"`
		} `json:"ticker"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for sym, d := range r.Data.Ticker {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Vol.V() * d.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.Buy.V() + d.Sell.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
