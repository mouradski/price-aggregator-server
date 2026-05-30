package woox

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// WOO X exposes a single bulk endpoint with 24h stats for every perpetual
// market. Symbols use the PERP_<BASE>_<QUOTE> format.
const futuresURL = "https://api.woox.io/v1/public/futures"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "woox" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Rows []struct {
		Symbol    string         `json:"symbol"`
		MarkPrice jsonutil.Float `json:"mark_price"`
		Close     jsonutil.Float `json:"24h_close"`
		Amount    jsonutil.Float `json:"24h_amount"` // 24h turnover (quote volume)
	} `json:"rows"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(futuresURL, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, row := range r.Rows {
		// symbol format: PERP_<BASE>_<QUOTE>
		parts := strings.Split(row.Symbol, "_")
		if len(parts) != 3 {
			continue
		}
		base, quote := parts[1], parts[2]
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		price := row.Close.V()
		if price == 0 {
			price = row.MarkPrice.V()
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: price, Timestamp: ts, H24Volume: row.Amount.V()})
	}
	return nil
}
