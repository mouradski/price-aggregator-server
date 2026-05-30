package aevo

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// Aevo returns every market in a single call. We keep only active perpetuals
// (options carry dated, multi-part instrument names) and use the mark price.
// The markets payload has no 24h volume field.
const url = "https://api.aevo.xyz/markets"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "aevo" }

func (c *Client) Interval() time.Duration { return 2 * time.Second }

type market struct {
	InstrumentType string         `json:"instrument_type"`
	Underlying     string         `json:"underlying_asset"`
	Quote          string         `json:"quote_asset"`
	MarkPrice      jsonutil.Float `json:"mark_price"`
	IsActive       bool           `json:"is_active"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var markets []market
	if err := client.GetJSON(url, &markets); err != nil {
		return err
	}
	ts := client.Now()
	for _, m := range markets {
		if m.InstrumentType != "PERPETUAL" || !m.IsActive {
			continue
		}
		if !ctx.IsAsset(m.Underlying) || !ctx.IsQuote(m.Quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: m.Underlying, Quote: m.Quote,
			LastPrice: m.MarkPrice.V(), Timestamp: ts, H24Volume: model.VolumeUnavailable})
	}
	return nil
}
