package nonkyc

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const url = "https://api.nonkyc.io/api/v2/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "nonkyc" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

type ticker struct {
	Base      string         `json:"base_currency"`
	Quote     string         `json:"target_currency"`
	LastPrice jsonutil.Float `json:"last_price"`
	USDVolume jsonutil.Float `json:"usd_volume_est"`
	Ask       jsonutil.Float `json:"ask"`
	Bid       jsonutil.Float `json:"bid"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var tickers []ticker
	if err := client.GetJSON(url, &tickers); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range tickers {
		if !ctx.IsAsset(t.Base) || !ctx.IsQuote(t.Quote) {
			continue
		}
		vol := t.USDVolume.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: t.Base, Quote: t.Quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: t.Base, Quote: t.Quote,
			LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
