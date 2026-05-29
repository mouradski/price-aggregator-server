package blofin

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const restURL = "https://openapi.blofin.com/api/v1/market/tickers"

// RestClient is the REST seed/fallback for Blofin, sharing the "blofin" name
// with the websocket Client.
type RestClient struct{}

func NewRest() *RestClient { return &RestClient{} }

func (c *RestClient) Name() string { return "blofin" }

func (c *RestClient) Interval() time.Duration { return 2 * time.Second }

type restResponse struct {
	Data []struct {
		InstID   string         `json:"instId"`
		Last     jsonutil.Float `json:"last"`
		AskPrice jsonutil.Float `json:"askPrice"`
		BidPrice jsonutil.Float `json:"bidPrice"`
		Vol24h   jsonutil.Float `json:"vol24h"`
	} `json:"data"`
}

func (c *RestClient) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r restResponse
	if err := client.GetJSON(restURL, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, d := range r.Data {
		base, quote := symbol.GetPair(d.InstID)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := d.Vol24h.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Last.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.AskPrice.V() + d.BidPrice.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
