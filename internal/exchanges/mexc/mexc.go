package mexc

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const (
	spotURL   = "https://api.mexc.com/api/v3/ticker/24hr"
	futureURL = "https://contract.mexc.com/api/v1/contract/ticker"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "mexc" }

func (c *Client) Interval() time.Duration { return time.Second }

type spotTicker struct {
	Symbol      string         `json:"symbol"`
	LastPrice   jsonutil.Float `json:"lastPrice"`
	AskPrice    jsonutil.Float `json:"askPrice"`
	BidPrice    jsonutil.Float `json:"bidPrice"`
	QuoteVolume jsonutil.Float `json:"quoteVolume"`
}

type contracts struct {
	Data []struct {
		Symbol    string         `json:"symbol"`
		LastPrice jsonutil.Float `json:"lastPrice"`
		Bid1      jsonutil.Float `json:"bid1"`
		Ask1      jsonutil.Float `json:"ask1"`
		Amount24  jsonutil.Float `json:"amount24"` // 24h turnover (quote volume)
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	ts := client.Now()

	var spot []spotTicker
	if err := client.GetJSON(spotURL, &spot); err != nil {
		return err
	}
	for _, t := range spot {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.QuoteVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.AskPrice.V() + t.BidPrice.V()) / 2, Timestamp: ts, H24Volume: vol})
	}

	var fut contracts
	if err := client.GetJSON(futureURL, &fut); err != nil {
		return nil // spot already succeeded; ignore futures failure
	}
	for _, d := range fut.Data {
		base, quote := symbol.GetPair(d.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Amount24.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "future", Base: base, Quote: quote,
			LastPrice: d.LastPrice.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "future-ask", Base: base, Quote: quote,
			LastPrice: (d.Ask1.V() + d.Bid1.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
