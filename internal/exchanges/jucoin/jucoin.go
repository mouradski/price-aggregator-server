package jucoin

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.jucoin.com/v1/spot/public/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "jucoin" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol  string         `json:"s"`
		Close   jsonutil.Float `json:"c"`
		Volume  jsonutil.Float `json:"v"`
		BestAsk jsonutil.Float `json:"ap"`
		BestBid jsonutil.Float `json:"bp"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, d := range r.Data {
		base, quote := symbol.GetPair(d.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Volume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Close.V(), Timestamp: ts, H24Volume: vol})
		if d.BestAsk.V() != 0 && d.BestBid.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (d.BestAsk.V() + d.BestBid.V()) / 2, Timestamp: ts, H24Volume: vol})
		}
	}
	return nil
}
