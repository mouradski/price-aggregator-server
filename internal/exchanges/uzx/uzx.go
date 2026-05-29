package uzx

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api-v2.uzx.com/notification/spot/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "uzx" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol string `json:"symbol"`
		Market struct {
			Close  jsonutil.Float `json:"close"`
			Volume jsonutil.Float `json:"turn_over"`
		} `json:"market"`
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
			LastPrice: t.Market.Close.V(), Timestamp: ts, H24Volume: t.Market.Volume.V()})
	}
	return nil
}
