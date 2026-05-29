package luno

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.luno.com/api/1/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "luno" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	Tickers []struct {
		Pair      string         `json:"pair"`
		LastTrade jsonutil.Float `json:"last_trade"`
		Bid       jsonutil.Float `json:"bid"`
		Ask       jsonutil.Float `json:"ask"`
		Volume24h jsonutil.Float `json:"rolling_24_hour_volume"` // base volume
	} `json:"tickers"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Tickers {
		base, quote := symbol.GetPair(t.Pair)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Volume24h.V() * t.LastTrade.V() // base -> quote volume
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastTrade.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
