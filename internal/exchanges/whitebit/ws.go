package whitebit

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

// WhiteBIT mainly lists USDT/USDC quotes; restrict the WS subscription to those
// (the REST fallback still covers any rarer quote).
var wsQuotes = []string{"USDT", "USDC"}

// WSClient streams WhiteBIT market updates over websocket, sharing the
// "whitebit" exchange name with the REST Client.
type WSClient struct{ id int }

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "whitebit" }

func (c *WSClient) URI(*client.Context) string { return "wss://api.whitebit.com/ws" }

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
	c.id++
	send(fmt.Sprintf(`{"id":%d,"method":"market_subscribe","params":[%s]}`, c.id, strings.Join(pairs, ",")))
}

func (c *WSClient) PingInterval() time.Duration { return 20 * time.Second }

func (c *WSClient) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"id":%d,"method":"ping","params":[]}`, c.id))
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "market_update") {
		return nil
	}
	var msg struct {
		Params []json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal([]byte(message), &msg); err != nil || len(msg.Params) < 2 {
		return nil
	}
	var market string
	if err := json.Unmarshal(msg.Params[0], &market); err != nil {
		return nil
	}
	var data struct {
		Last jsonutil.Float `json:"last"`
		Deal jsonutil.Float `json:"deal"`
	}
	if err := json.Unmarshal(msg.Params[1], &data); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(market)
	// Match the REST client, which emits this exchange under the "-ask" name.
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name() + "-ask", Base: base, Quote: quote,
		LastPrice: data.Last.V(), H24Volume: data.Deal.V(),
	}}
}
