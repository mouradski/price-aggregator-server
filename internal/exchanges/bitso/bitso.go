package bitso

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://bitso.com/api/v3/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitso" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	Payload []struct {
		Book   string         `json:"book"`
		Last   jsonutil.Float `json:"last"`
		Volume jsonutil.Float `json:"volume"`
		Ask    jsonutil.Float `json:"ask"`
		Bid    jsonutil.Float `json:"bid"`
	} `json:"payload"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, m := range r.Payload {
		base, quote := symbol.GetPair(m.Book)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := m.Volume.V() * m.Last.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: m.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (m.Ask.V() + m.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
