package bitunix

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://fapi.bitunix.com/api/v1/futures/market/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitunix" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol   string         `json:"symbol"`
		Last     jsonutil.Float `json:"last"`
		QuoteVol jsonutil.Float `json:"quoteVol"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: t.QuoteVol.V()})
	}
	return nil
}
