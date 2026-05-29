package bit2me

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://gateway.bit2me.com/v2/trading/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bit2me" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type ticker struct {
	Symbol      string         `json:"symbol"`
	Close       jsonutil.Float `json:"close"`
	Ask         jsonutil.Float `json:"ask"`
	Bid         jsonutil.Float `json:"bid"`
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
		vol := t.QuoteVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Close.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
