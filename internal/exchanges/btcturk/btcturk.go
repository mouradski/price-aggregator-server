package btcturk

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.btcturk.com/api/v2/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "btcturk" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Pair   string         `json:"pair"`
		Last   jsonutil.Float `json:"last"`
		Ask    jsonutil.Float `json:"ask"`
		Bid    jsonutil.Float `json:"bid"`
		Volume jsonutil.Float `json:"volume"`
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
		vol := t.Volume.V() * t.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
