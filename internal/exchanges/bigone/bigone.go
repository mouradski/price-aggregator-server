package bigone

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://big.one/api/v3/asset_pairs/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bigone" }

func (c *Client) Interval() time.Duration { return time.Second }

type priceDetail struct {
	Price jsonutil.Float `json:"price"`
}

type response struct {
	Data []struct {
		AssetPairName string         `json:"asset_pair_name"`
		Bid           priceDetail    `json:"bid"`
		Ask           priceDetail    `json:"ask"`
		Close         jsonutil.Float `json:"close"`
		Volume        jsonutil.Float `json:"volume"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, d := range r.Data {
		base, quote := symbol.GetPair(d.AssetPairName)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Volume.V() * d.Close.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Close.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.Ask.Price.V() + d.Bid.Price.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
