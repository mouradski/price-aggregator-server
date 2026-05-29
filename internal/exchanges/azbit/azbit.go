package azbit

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://data.azbit.com/api/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "azbit" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

type ticker struct {
	Symbol    string         `json:"currencyPairCode"`
	Price     jsonutil.Float `json:"price"`
	AskPrice  jsonutil.Float `json:"askPrice"`
	BidPrice  jsonutil.Float `json:"bidPrice"`
	Volume24h jsonutil.Float `json:"volume24h"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		if t.Price.V() == 0 {
			continue
		}
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Volume24h.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Price.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.AskPrice.V() + t.BidPrice.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
