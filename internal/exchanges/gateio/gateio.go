package gateio

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.gateio.ws/api/v4/spot/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "gateio" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type ticker struct {
	CurrencyPair string         `json:"currency_pair"`
	Ask          jsonutil.Float `json:"lowest_ask"`
	Bid          jsonutil.Float `json:"highest_bid"`
	Last         jsonutil.Float `json:"last"`
	QuoteVolume  jsonutil.Float `json:"quote_volume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		base, quote := symbol.GetPair(t.CurrencyPair)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.QuoteVolume.V()
		push(model.Ticker{
			Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol,
		})
		if t.Ask.V() != 0 && t.Bid.V() != 0 {
			push(model.Ticker{
				Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol,
			})
		}
	}
	return nil
}
