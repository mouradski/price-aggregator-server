package xt

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://sapi.xt.com/v4/public/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "xt" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Result []struct {
		S  string         `json:"s"`
		C  jsonutil.Float `json:"c"`
		Ap jsonutil.Float `json:"ap"`
		Bp jsonutil.Float `json:"bp"`
		V  jsonutil.Float `json:"v"`
	} `json:"result"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Result {
		base, quote := symbol.GetPair(t.S)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.V.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.C.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ap.V() + t.Bp.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
