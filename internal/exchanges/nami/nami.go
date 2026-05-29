package nami

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const url = "https://nami.exchange/api/v1.0/market/summaries"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "nami" }

func (c *Client) Interval() time.Duration { return time.Second }

// The "symbol" field is unreliable (some entries return the number 0), so the
// pair is built from base_currency / exchange_currency instead.
type response struct {
	Data []struct {
		LastPrice        jsonutil.Float  `json:"last_price"`
		BaseCurrency     jsonutil.String `json:"base_currency"`     // quote currency (e.g. USDT)
		ExchangeCurrency jsonutil.String `json:"exchange_currency"` // base currency (e.g. ETH)
		TotalBaseVolume  jsonutil.Float  `json:"total_base_volume"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Data {
		base := strings.ToUpper(t.ExchangeCurrency.V())
		quote := strings.ToUpper(t.BaseCurrency.V())
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: t.TotalBaseVolume.V()})
	}
	return nil
}
