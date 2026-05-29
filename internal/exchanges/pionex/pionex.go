package pionex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.pionex.com/api/v1/market/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "pionex" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data struct {
		Tickers []struct {
			Symbol string         `json:"symbol"`
			Close  jsonutil.Float `json:"close"`
			Amount jsonutil.Float `json:"amount"`
		} `json:"tickers"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data.Tickers {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Close.V(), Timestamp: ts, H24Volume: t.Amount.V()})
	}
	return nil
}
