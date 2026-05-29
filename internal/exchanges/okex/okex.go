package okex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://www.okx.com/api/v5/market/tickers?instType=SPOT"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "okex" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		InstID    string         `json:"instId"`
		Last      jsonutil.Float `json:"last"`
		BidPx     jsonutil.Float `json:"bidPx"`
		AskPx     jsonutil.Float `json:"askPx"`
		VolCcy24h jsonutil.Float `json:"volCcy24h"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data {
		base, quote := symbol.GetPair(t.InstID)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.VolCcy24h.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.BidPx.V() + t.AskPx.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
