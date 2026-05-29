package cointr

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

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "cointr" }

func (c *Client) URI(*client.Context) string { return "wss://ws.cointr.com/v2/ws/public" }

// CoinTR lists USDT/USDC markets; one non-existent instId in the batch breaks
// the whole subscription, so restrict to those quotes.
var wsQuotes = []string{"USDT", "USDC"}

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var args []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			args = append(args, fmt.Sprintf(`{"instType":"SPOT","channel":"ticker","instId":"%s%s"}`, base, quote))
		}
	}
	send(fmt.Sprintf(`{"op":"subscribe","args":[%s]}`, strings.Join(args, ",")))
}

func (c *Client) PingInterval() time.Duration { return 30 * time.Second }

func (c *Client) Ping(send func(string)) { send("ping") }

type response struct {
	Data []struct {
		LastPr      jsonutil.Float `json:"lastPr"`
		AskPr       jsonutil.Float `json:"askPr"`
		BidPr       jsonutil.Float `json:"bidPr"`
		InstID      string         `json:"instId"`
		QuoteVolume jsonutil.Float `json:"quoteVolume"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "snapshot") {
		return nil
	}
	var r response
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range r.Data {
		base, quote := symbol.GetPair(d.InstID)
		vol := d.QuoteVolume.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: d.LastPr.V(), H24Volume: vol},
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (d.AskPr.V() + d.BidPr.V()) / 2, H24Volume: vol})
	}
	return out
}
