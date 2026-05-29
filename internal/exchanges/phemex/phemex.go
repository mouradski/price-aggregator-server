package phemex

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const (
	url   = "https://api.phemex.com/md/spot/ticker/24hr/all"
	scale = 100000000.0 // Phemex prices/volumes are integer-scaled by 1e8.
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "phemex" }

func (c *Client) Interval() time.Duration { return time.Second }

type response struct {
	Result []struct {
		Symbol     string         `json:"symbol"`
		LastEp     jsonutil.Float `json:"lastEp"`
		AskEp      jsonutil.Float `json:"askEp"`
		BidEp      jsonutil.Float `json:"bidEp"`
		TurnoverEv jsonutil.Float `json:"turnoverEv"`
	} `json:"result"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r response
	if err := client.GetJSON(url, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Result {
		if t.Symbol == "" {
			continue
		}
		base, quote := symbol.GetPair(strings.TrimPrefix(t.Symbol, "s"))
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.TurnoverEv.V() / scale
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastEp.V() / scale, Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.AskEp.V() + t.BidEp.V()) / (2 * scale), Timestamp: ts, H24Volume: vol})
	}
	return nil
}
