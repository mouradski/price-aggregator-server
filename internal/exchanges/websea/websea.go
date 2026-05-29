package websea

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://oapi.websea.com/openApi/market/24kline"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "websea" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Result []struct {
		Symbol string `json:"symbol"`
		Data   struct {
			Close jsonutil.Float `json:"close"`
			Vol   jsonutil.Float `json:"vol"`
		} `json:"data"`
		Ask jsonutil.Float `json:"ask"`
		Bid jsonutil.Float `json:"bid"`
	} `json:"result"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Result {
		base, quote := symbol.GetPair(t.Symbol)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Data.Vol.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.Data.Close.V(), Timestamp: ts, H24Volume: vol})
		if t.Ask.V() != 0 && t.Bid.V() != 0 {
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
		}
	}
	return nil
}
