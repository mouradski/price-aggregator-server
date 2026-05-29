package coinbase

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.coinbase.com/api/v3/brokerage/market/products?product_type=SPOT"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "coinbase" }

func (c *Client) Interval() time.Duration { return time.Second }

type products struct {
	Products []struct {
		Price     jsonutil.Float `json:"price"`
		ProductID string         `json:"product_id"`
		Volume24h jsonutil.Float `json:"volume_24h"`
	} `json:"products"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var resp products
	if err := client.GetJSON(url, &resp); err != nil {
		return err
	}
	for _, p := range resp.Products {
		base, quote := symbol.GetPair(p.ProductID)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{
			Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: p.Price.V(), Timestamp: client.Now(), H24Volume: p.Volume24h.V() * p.Price.V(),
		})
	}
	return nil
}
