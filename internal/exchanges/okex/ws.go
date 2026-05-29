package okex

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

// WSClient streams OKX spot tickers over websocket. It shares the "okex"
// exchange name with the REST Client.
type WSClient struct{}

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "okex" }

func (c *WSClient) URI(*client.Context) string { return "wss://ws.okx.com:8443/ws/v5/public" }

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var args []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			args = append(args, fmt.Sprintf(`{"channel":"tickers","instId":"%s-%s"}`, base, quote))
		}
	}
	// OKX accepts many args per subscribe; chunk to stay within message limits.
	const chunk = 100
	for i := 0; i < len(args); i += chunk {
		end := i + chunk
		if end > len(args) {
			end = len(args)
		}
		send(`{"op":"subscribe","args":[` + strings.Join(args[i:end], ",") + `]}`)
	}
}

func (c *WSClient) PingInterval() time.Duration { return 25 * time.Second }

func (c *WSClient) Ping(send func(string)) { send("ping") }

type wsResponse struct {
	Data []struct {
		InstID    string         `json:"instId"`
		Last      jsonutil.Float `json:"last"`
		BidPx     jsonutil.Float `json:"bidPx"`
		AskPx     jsonutil.Float `json:"askPx"`
		VolCcy24h jsonutil.Float `json:"volCcy24h"`
	} `json:"data"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"tickers"`) || !strings.Contains(message, "last") {
		return nil
	}
	var r wsResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, t := range r.Data {
		base, quote := symbol.GetPair(t.InstID)
		vol := t.VolCcy24h.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: t.Last.V(), H24Volume: vol},
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (t.BidPx.V() + t.AskPx.V()) / 2, H24Volume: vol})
	}
	return out
}
