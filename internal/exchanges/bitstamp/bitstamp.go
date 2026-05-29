package bitstamp

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://www.bitstamp.net/api/v2/ticker/"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitstamp" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	Last   jsonutil.Float `json:"last"`
	Ask    jsonutil.Float `json:"ask"`
	Bid    jsonutil.Float `json:"bid"`
	Pair   string         `json:"pair"`
	Volume jsonutil.Float `json:"volume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		base, quote := symbol.GetPair(t.Pair)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		volume := t.Volume.V() * t.Last.V()
		push(model.Ticker{
			Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: volume,
		})
		push(model.Ticker{
			Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: volume,
		})
	}
	return nil
}
