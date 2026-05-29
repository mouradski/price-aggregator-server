package bitpanda

import (
	"strconv"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
)

const url = "https://api.bitpanda.com/v1/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitpanda" }

func (c *Client) Interval() time.Duration { return time.Second }

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var data map[string]map[string]string
	if err := client.GetJSON(url, &data); err != nil {
		return err
	}
	ts := client.Now()
	for base, quotes := range data {
		if !ctx.IsAsset(base) {
			continue
		}
		for quote, priceStr := range quotes {
			if !ctx.IsQuote(quote) {
				continue
			}
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				continue
			}
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: price, Timestamp: ts, H24Volume: model.VolumeUnavailable})
		}
	}
	return nil
}
