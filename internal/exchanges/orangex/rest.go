package orangex

import (
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// REST seed/fallback for OrangeX spot, sharing the "orangex" name with the WS
// Client. instrument_name here is "BASE-QUOTE" (no -SPOT suffix).
const restURL = "https://api.orangex.com/api/v1/public/tickers?currency=SPOT"

type RestClient struct{}

func NewRest() *RestClient { return &RestClient{} }

func (c *RestClient) Name() string { return "orangex" }

func (c *RestClient) Interval() time.Duration { return 2 * time.Second }

type restResponse struct {
	Result []struct {
		InstrumentName string         `json:"instrument_name"`
		LastPrice      jsonutil.Float `json:"last_price"`
		BestBidPrice   jsonutil.Float `json:"best_bid_price"`
		BestAskPrice   jsonutil.Float `json:"best_ask_price"`
		Stats          struct {
			Turnover jsonutil.Float `json:"turnover"` // 24h quote volume
		} `json:"stats"`
	} `json:"result"`
}

func (c *RestClient) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var r restResponse
	if err := client.GetJSON(restURL, &r); err != nil {
		return err
	}
	ts := client.Now()
	for _, t := range r.Result {
		base, quote := symbol.GetPair(t.InstrumentName)
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		vol := t.Stats.Turnover.V()
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: t.LastPrice.V(), Timestamp: ts, H24Volume: vol})
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (t.BestBidPrice.V() + t.BestAskPrice.V()) / 2, Timestamp: ts, H24Volume: vol})
	}
	return nil
}
