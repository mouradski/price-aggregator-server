package bibox

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const url = "https://api.bibox.com/v3/mdata/marketAll"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bibox" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Result []struct {
		Base   string         `json:"coin_symbol"`
		Quote  string         `json:"currency_symbol"`
		Last   jsonutil.Float `json:"last"`
		Vol24H jsonutil.Float `json:"vol24H"`
	} `json:"result"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Result {
		if !ctx.IsAsset(t.Base) || !ctx.IsQuote(t.Quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: t.Base, Quote: t.Quote,
			LastPrice: t.Last.V(), Timestamp: ts, H24Volume: t.Vol24H.V() * t.Last.V()})
	}
	return nil
}
