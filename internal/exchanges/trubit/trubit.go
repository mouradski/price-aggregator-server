package trubit

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api-spot.trubit.com/openapi/quote/v1/ticker/24hr"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "trubit" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	Symbol      string         `json:"symbol"`
	LastPrice   jsonutil.Float `json:"lastPrice"`
	QuoteVolume jsonutil.Float `json:"quoteVolume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: t.QuoteVolume.V()})
	}
	return nil
}
