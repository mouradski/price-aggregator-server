package gleec

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.exchange.gleec.com/api/3/public/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "gleec" }

func (c *Client) Interval() time.Duration { return time.Second }

type ticker struct {
	Ask    jsonutil.Float `json:"ask"`
	Bid    jsonutil.Float `json:"bid"`
	Last   jsonutil.Float `json:"last"`
	Volume jsonutil.Float `json:"volume_quote"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers map[string]ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for sym, t := range tickers {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Volume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
		if t.Ask.V() != 0 && t.Bid.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
		}
	}
	return nil
}
