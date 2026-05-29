package cryptomus

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.cryptomus.com/v1/exchange/market/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "cryptomus" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol    string         `json:"currency_pair"`
		LastPrice jsonutil.Float `json:"last_price"`
		Volume    jsonutil.Float `json:"quote_volume"`
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
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: t.Volume.V()})
	}
	return nil
}
