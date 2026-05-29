package icrypex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.icrypex.com/v1/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "icrypex" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	Last   jsonutil.Float `json:"last"`
	Ask    jsonutil.Float `json:"ask"`
	Bid    jsonutil.Float `json:"bid"`
	Volume jsonutil.Float `json:"volume"`
	Symbol string         `json:"symbol"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Volume.V() * t.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Bid.V() + t.Ask.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
