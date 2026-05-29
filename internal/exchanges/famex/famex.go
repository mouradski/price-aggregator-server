package famex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const url = "https://api.fameex.com/v2/public/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "famex" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Data map[string]struct {
		Base        string         `json:"base_id"`
		Quote       string         `json:"quote_id"`
		LastPrice   jsonutil.Float `json:"last_price"`
		QuoteVolume jsonutil.Float `json:"quote_volume"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, m := range r.Data {
		if !ctx.IsAsset(m.Base) || !ctx.IsQuote(m.Quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: m.Base, Quote: m.Quote,
			LastPrice: m.LastPrice.V(), Timestamp: ts, H24Volume: m.QuoteVolume.V()})
	}
	return nil
}
