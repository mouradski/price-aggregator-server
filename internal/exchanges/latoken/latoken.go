package latoken

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.latoken.com/v2/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "latoken" }

func (c *Client) Interval() time.Duration { return 5 * time.Second }

type ticker struct {
	Symbol    string         `json:"symbol"`
	LastPrice jsonutil.Float `json:"lastPrice"`
	BestBid   jsonutil.Float `json:"bestBid"`
	BestAsk   jsonutil.Float `json:"bestAsk"`
	Volume24h jsonutil.Float `json:"volume24h"`
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
		vol := t.Volume24h.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.BestAsk.V() + t.BestBid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
