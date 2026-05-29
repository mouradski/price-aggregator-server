package btse

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.btse.com/spot/api/v3.2/market_summary"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "btse" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type ticker struct {
	Symbol     string         `json:"symbol"`
	Last       jsonutil.Float `json:"last"`
	LowestAsk  jsonutil.Float `json:"lowest_ask"`
	HighestBid jsonutil.Float `json:"highest_bid"`
	Volume     jsonutil.Float `json:"volume"`
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
		vol := t.Volume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.LowestAsk.V() + t.HighestBid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
