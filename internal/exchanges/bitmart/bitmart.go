package bitmart

import (
	"strconv"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api-cloud.bitmart.com/spot/quotation/v3/tickers"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "bitmart" }

func (c *Client) Interval() time.Duration { return time.Second }

// Bitmart returns each ticker as a positional string array:
// [symbol, last, _, volume, _, _, _, _, ask, _, bid, ...]
type response struct {
	Data [][]string `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, row := range r.Data {
		if len(row) < 11 {
			continue
		}
		base, quote := symbol.GetPair(row[0])
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := parse(row[3])
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: parse(row[1]), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (parse(row[8]) + parse(row[10])) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}

func parse(s string) float64 {
	v, _ := strconv.ParseFloat(s, 64)
	return v
}
