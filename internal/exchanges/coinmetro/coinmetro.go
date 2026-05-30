package coinmetro

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// Coinmetro returns the latest price for every pair in a single call. The
// payload carries base/quote split out, plus best ask/bid, but no 24h volume.
const url = "https://api.coinmetro.com/exchange/prices"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "coinmetro" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	LatestPrices []struct {
		Base  string         `json:"base"`
		Quote string         `json:"quote"`
		Price jsonutil.Float `json:"price"`
		Ask   jsonutil.Float `json:"ask"`
		Bid   jsonutil.Float `json:"bid"`
	} `json:"latestPrices"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, p := range r.LatestPrices {
		if !ctx.IsAsset(p.Base) || !ctx.IsQuote(p.Quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: p.Base, Quote: p.Quote,
			LastPrice: p.Price.V(), Timestamp: ts, H24Volume: model.VolumeUnavailable})
		if p.Ask.V() != 0 && p.Bid.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: p.Base, Quote: p.Quote,
				LastPrice: (p.Ask.V() + p.Bid.V()) / 2, Timestamp: ts, H24Volume: model.VolumeUnavailable})
		}
	}
	return nil
}
