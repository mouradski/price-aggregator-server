package bitget

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.bitget.com/api/v2/spot/market/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitget" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data []struct {
		Symbol      string         `json:"symbol"`
		LastPr      jsonutil.Float `json:"lastPr"`
		AskPr       jsonutil.Float `json:"askPr"`
		BidPr       jsonutil.Float `json:"bidPr"`
		QuoteVolume jsonutil.Float `json:"quoteVolume"`
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
		vol := t.QuoteVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPr.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.AskPr.V() + t.BidPr.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
