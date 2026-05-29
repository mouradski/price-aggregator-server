package coinbase

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Coinbase quotes its markets against these among the configured quotes.
var wsQuotes = []string{"USD", "USDC", "USDT"}

// WSClient streams Coinbase Exchange tickers over websocket, sharing the
// "coinbase" exchange name with the REST Client.
type WSClient struct{}

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "coinbase" }

func (c *WSClient) URI(*client.Context) string { return "wss://ws-feed.exchange.coinbase.com" }

func (c *WSClient) Subscribe(ctx *client.Context, send func(string)) {
	var products []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if !ctx.IsQuote(quote) {
				continue
			}
			products = append(products, `"`+base+"-"+quote+`"`)
		}
	}
	send(`{"type":"subscribe","product_ids":[` + strings.Join(products, ",") + `],"channels":["ticker"]}`)
}

type wsTicker struct {
	Type      string         `json:"type"`
	ProductID string         `json:"product_id"`
	Price     jsonutil.Float `json:"price"`
	Volume24h jsonutil.Float `json:"volume_24h"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "product_id") || !strings.Contains(message, `"ticker"`) {
		return nil
	}
	var t wsTicker
	if err := json.Unmarshal([]byte(message), &t); err != nil || t.ProductID == "" {
		return nil
	}
	base, quote := symbol.GetPair(t.ProductID)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: t.Price.V(), H24Volume: t.Volume24h.V() * t.Price.V(),
	}}
}
