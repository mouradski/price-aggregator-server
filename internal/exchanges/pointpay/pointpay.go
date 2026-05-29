package pointpay

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.pointpay.io/api/v1/public/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "pointpay" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Result map[string]struct {
		Ticker struct {
			Bid  jsonutil.Float `json:"bid"`
			Ask  jsonutil.Float `json:"ask"`
			Last jsonutil.Float `json:"last"`
			Vol  jsonutil.Float `json:"vol"`
		} `json:"ticker"`
	} `json:"result"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for sym, v := range r.Result {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		t := v.Ticker
		vol := t.Vol.V() * t.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
