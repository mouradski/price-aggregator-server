package digifinex

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

// Digifinex lists mostly USDT/USDC markets; restrict the WS subscription to
// those (REST fallback covers any rarer quote). gzip-compressed frames.
var wsQuotes = []string{"USDT", "USDC"}

type WSClient struct{ id int }

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "digifinex" }

func (c *WSClient) URI(*client.Context) string { return "wss://openapi.digifinex.com/ws/v1/" }

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.Assets() { // lowercase
		for _, quote := range wsQuotes {
			lq := strings.ToLower(quote)
			if base == lq || !ctx.IsQuote(quote) {
				continue
			}
			pairs = append(pairs, `"`+base+"_"+lq+`"`)
		}
	}
	c.id++
	send(fmt.Sprintf(`{"id":%d,"method":"ticker.subscribe","params":[%s]}`, c.id, strings.Join(pairs, ",")))
}

func (c *WSClient) PingInterval() time.Duration { return 25 * time.Second }

func (c *WSClient) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"id":%d,"method":"server.ping","params":[]}`, c.id))
}

type wsResponse struct {
	Method string `json:"method"`
	Params []struct {
		Symbol        string         `json:"symbol"`
		Last          jsonutil.Float `json:"last"`
		QuoteVolume24 jsonutil.Float `json:"quote_volume_24h"`
		BestBid       jsonutil.Float `json:"best_bid"`
		BestAsk       jsonutil.Float `json:"best_ask"`
	} `json:"params"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker.update") {
		return nil
	}
	var r wsResponse
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range r.Params {
		base, quote := symbol.GetPair(d.Symbol)
		vol := d.QuoteVolume24.V()
		out = append(out,
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: d.Last.V(), H24Volume: vol},
			model.Ticker{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
				LastPrice: (d.BestAsk.V() + d.BestBid.V()) / 2, H24Volume: vol})
	}
	return out
}
