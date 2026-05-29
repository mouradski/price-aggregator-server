package gateio

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

// WSClient streams Gate.io spot tickers over websocket, sharing the "gateio"
// exchange name with the REST Client.
type WSClient struct{}

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "gateio" }

func (c *WSClient) URI(*client.Context) string { return "wss://api.gateio.ws/ws/v4/" }

// Gate.io rejects an entire subscription batch if it contains a non-existent
// pair, so the WS feed only subscribes to quotes Gate actually lists; the REST
// fallback still covers any rarer quote.
var wsQuotes = []string{"USDT", "USDC"}

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			pairs = append(pairs, `"`+base+"_"+quote+`"`)
		}
	}
	send(fmt.Sprintf(`{"time":%d,"channel":"spot.tickers","event":"subscribe","payload":[%s]}`,
		time.Now().Unix(), strings.Join(pairs, ",")))
}

func (c *WSClient) PingInterval() time.Duration { return 20 * time.Second }

func (c *WSClient) Ping(send func(string)) {
	send(fmt.Sprintf(`{"time":%d,"channel":"spot.ping"}`, time.Now().Unix()))
}

type wsResponse struct {
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Result  struct {
		CurrencyPair string         `json:"currency_pair"`
		Last         jsonutil.Float `json:"last"`
		LowestAsk    jsonutil.Float `json:"lowest_ask"`
		HighestBid   jsonutil.Float `json:"highest_bid"`
		QuoteVolume  jsonutil.Float `json:"quote_volume"`
	} `json:"result"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "spot.tickers") || !strings.Contains(message, "update") {
		return nil
	}
	var r wsResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil || r.Result.CurrencyPair == "" {
		return nil
	}
	base, quote := symbol.GetPair(r.Result.CurrencyPair)
	vol := r.Result.QuoteVolume.V()
	ts := client.Now()
	out := []model.Ticker{{
		Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: r.Result.Last.V(), H24Volume: vol,
	}}
	if r.Result.LowestAsk.V() != 0 && r.Result.HighestBid.V() != 0 {
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (r.Result.LowestAsk.V() + r.Result.HighestBid.V()) / 2, H24Volume: vol,
		})
	}
	return out
}
