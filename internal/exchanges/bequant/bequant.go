package bequant

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bequant" }

func (c *Client) URI(*client.Context) string { return "wss://api.bequant.io/api/3/ws/public" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var bases []string
	for _, b := range ctx.AssetsUpper() {
		bases = append(bases, `"`+b+`"`)
	}
	csv := strings.Join(bases, ",")
	for _, quote := range ctx.QuotesUpper() {
		c.id++
		send(fmt.Sprintf(`{"method":"subscribe","ch":"price/rate/1s","params":{"currencies":[%s],"target_currency":"%s"},"id":%d}`,
			csv, quote, c.id))
	}
}

type cryptoRate struct {
	TargetCurrency string `json:"target_currency"`
	Data           map[string]struct {
		R jsonutil.Float `json:"r"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "price/rate/1s") {
		return nil
	}
	var r cryptoRate
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for base, d := range r.Data {
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: r.TargetCurrency,
			LastPrice: d.R.V(), H24Volume: model.VolumeUnavailable,
		})
	}
	return out
}
