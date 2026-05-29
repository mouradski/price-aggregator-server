package lbank

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.lbkex.com/v2/ticker/24hr.do?symbol=all"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "lbank" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol string `json:"symbol"`
		Ticker struct {
			Latest jsonutil.Float `json:"latest"`
			Vol    jsonutil.Float `json:"vol"`
		} `json:"ticker"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, e := range r.Data {
		base, quote := symbol.GetPair(e.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: e.Ticker.Latest.V(), Timestamp: ts, H24Volume: e.Ticker.Vol.V() * e.Ticker.Latest.V()})
	}
	return nil
}
