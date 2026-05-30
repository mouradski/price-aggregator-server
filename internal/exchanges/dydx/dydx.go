package dydx

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// dYdX v4 exposes every perpetual market (keyed by "<BASE>-USD") with its
// oracle price and 24h quote volume in a single indexer call.
const url = "https://indexer.dydx.trade/v4/perpetualMarkets"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "dydx" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Markets map[string]struct {
		Ticker      string         `json:"ticker"`
		OraclePrice jsonutil.Float `json:"oraclePrice"`
		Volume24H   jsonutil.Float `json:"volume24H"`
	} `json:"markets"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, m := range r.Markets {
		// ticker format: <BASE>-USD
		base, quote, ok := strings.Cut(m.Ticker, "-")
		if !ok || !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: m.OraclePrice.V(), Timestamp: ts, H24Volume: m.Volume24H.V()})
	}
	return nil
}
