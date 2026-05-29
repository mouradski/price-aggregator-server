package whitebit

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://whitebit.com/api/v4/public/ticker"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "whitebit" }

func (c *Client) Interval() time.Duration { return time.Second }

type entry struct {
	LastPrice   jsonutil.Float `json:"last_price"`
	QuoteVolume jsonutil.Float `json:"quote_volume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var markets map[string]entry
	if err := client.GetJSON(url, &markets); err != nil {
		return err
	}
	ts := client.Now()
	for market, d := range markets {
		base, quote := symbol.GetPair(strings.ReplaceAll(market, "_", ""))
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{
			Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: d.LastPrice.V(), Timestamp: ts, H24Volume: d.QuoteVolume.V(),
		})
	}
	return nil
}
