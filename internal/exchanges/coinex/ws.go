package coinex

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

// CoinEx v2 lists mostly USDT/USDC markets; the WS state channel does not carry
// best bid/ask, so the WS feed only emits the main ticker (the "-ask" variant
// stays available through the REST fallback). gzip-compressed frames.
var wsQuotes = []string{"USDT", "USDC"}

type WSClient struct{ id int }

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "coinex" }

func (c *WSClient) URI(*client.Context) string { return "wss://socket.coinex.com/v2/spot" }

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var markets []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			markets = append(markets, `"`+base+quote+`"`)
		}
	}
	c.id++
	send(fmt.Sprintf(`{"method":"state.subscribe","params":{"market_list":[%s]},"id":%d}`,
		strings.Join(markets, ","), c.id))
}

func (c *WSClient) PingInterval() time.Duration { return 25 * time.Second }

func (c *WSClient) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"method":"server.ping","params":{},"id":%d}`, c.id))
}

type wsResponse struct {
	Method string `json:"method"`
	Data   struct {
		StateList []struct {
			Market string         `json:"market"`
			Last   jsonutil.Float `json:"last"`
			Value  jsonutil.Float `json:"value"` // quote volume
		} `json:"state_list"`
	} `json:"data"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "state.update") {
		return nil
	}
	var r wsResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, s := range r.Data.StateList {
		base, quote := symbol.GetPair(s.Market)
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: s.Last.V(), H24Volume: s.Value.V(),
		})
	}
	return out
}
