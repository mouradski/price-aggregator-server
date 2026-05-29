package bullish

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const url = "https://api.exchange.bullish.com/aggregator-api/v1/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bullish" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	BaseSymbol  string         `json:"baseSymbol"`
	QuoteSymbol string         `json:"quoteSymbol"`
	Last        jsonutil.Float `json:"last"`
	BestAsk     jsonutil.Float `json:"bestAsk"`
	BestBid     jsonutil.Float `json:"bestBid"`
	QuoteVolume jsonutil.Float `json:"quoteVolume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers map[string]ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for key, t := range tickers {
		if len(key) >= 12 {
			continue
		}
		if !ctx.IsAsset(t.BaseSymbol) || !ctx.IsQuote(t.QuoteSymbol) {
			continue
		}
		vol := t.QuoteVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: t.BaseSymbol, Quote: t.QuoteSymbol,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		if t.BestAsk.V() != 0 && t.BestBid.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: t.BaseSymbol, Quote: t.QuoteSymbol,
				LastPrice: (t.BestAsk.V() + t.BestBid.V()) / 2, Timestamp: ts, H24Volume: vol})
		}
	}
	return nil
}
