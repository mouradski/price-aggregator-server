package bitmart

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type WSClient struct{}

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "bitmart" }

func (c *WSClient) URI(*client.Context) string {
	return "wss://ws-manager-compress.bitmart.com/api?protocol=1.1"
}

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var args []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			args = append(args, fmt.Sprintf(`"spot/ticker:%s_%s"`, base, quote))
		}
	}
	send(`{"op":"subscribe","args":[` + strings.Join(args, ",") + `]}`)
}

func (c *WSClient) PingInterval() time.Duration { return 15 * time.Second }

func (c *WSClient) Ping(send func(string)) { send("ping") }

type wsResponse struct {
	Table string `json:"table"`
	Data  []struct {
		Symbol        string         `json:"symbol"`
		LastPrice     jsonutil.Float `json:"last_price"`
		AskPx         jsonutil.Float `json:"ask_px"`
		BidPx         jsonutil.Float `json:"bid_px"`
		QuoteVolume24 jsonutil.Float `json:"quote_volume_24h"`
	} `json:"data"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "spot/ticker") || !strings.Contains(message, "last_price") {
		return nil
	}
	var r wsResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range r.Data {
		base, quote := symbol.GetPair(d.Symbol)
		vol := d.QuoteVolume24.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: d.LastPrice.V(), H24Volume: vol},
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (d.AskPx.V() + d.BidPx.V()) / 2, H24Volume: vol})
	}
	return out
}
