package cexio

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// CEX.IO has no public websocket *ticker* stream (the new API requires auth, the
// old API's ticker call is one-shot). Its only public live feed is the order
// book ("pair-BASE-QUOTE" rooms), so the price is taken as the mid of the best
// bid/ask from each full `md` snapshot (pushed every few seconds). This is a
// WS-only exchange (no REST) to avoid the REST rate limit.
type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "cexio" }

func (c *Client) URI(*client.Context) string { return "wss://ws.cex.io/ws/" }

// CEX.IO order books are quoted against these among the configured quotes.
var wsQuotes = []string{"USD", "USDT", "USDC"}

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var rooms []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			rooms = append(rooms, `"pair-`+base+"-"+quote+`"`)
		}
	}
	send(`{"e":"subscribe","rooms":[` + strings.Join(rooms, ",") + `]}`)
}

// Pong answers CEX.IO's heartbeat ({"e":"ping"} -> {"e":"pong"}).
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, `"ping"`) {
		send(`{"e":"pong"}`)
		return true
	}
	return false
}

type mdMessage struct {
	Data struct {
		Pair string      `json:"pair"`
		Buy  [][]float64 `json:"buy"`  // [price, amount], best bid first
		Sell [][]float64 `json:"sell"` // [price, amount], best ask first
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"e":"md"`) {
		return nil
	}
	var m mdMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	if len(m.Data.Buy) == 0 || len(m.Data.Sell) == 0 ||
		len(m.Data.Buy[0]) < 1 || len(m.Data.Sell[0]) < 1 {
		return nil
	}
	base, quote := symbol.GetPair(m.Data.Pair) // "BTC:USD"
	mid := (m.Data.Buy[0][0] + m.Data.Sell[0][0]) / 2
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: mid, H24Volume: model.VolumeUnavailable,
	}}
}
