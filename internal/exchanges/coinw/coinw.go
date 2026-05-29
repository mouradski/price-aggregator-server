package coinw

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.coinw.com/api/v1/public?command=returnTicker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "coinw" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type response struct {
	Data map[string]struct {
		Last       jsonutil.Float `json:"last"`
		HighestBid jsonutil.Float `json:"highestBid"`
		LowestAsk  jsonutil.Float `json:"lowestAsk"`
		BaseVolume jsonutil.Float `json:"baseVolume"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for sym, d := range r.Data {
		base, quote := symbol.GetPair(sym)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.BaseVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.LowestAsk.V() + d.HighestBid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
