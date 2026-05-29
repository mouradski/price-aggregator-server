package bybit

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

func (c *Client) Name() string { return "bybit" }

func (c *Client) URI(*client.Context) string { return "wss://stream.bybit.com/v5/public/spot" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			send(fmt.Sprintf(`{"op":"subscribe","args":["tickers.%s%s"]}`, base, quote))
		}
	}
}

func (c *Client) PingInterval() time.Duration { return 30 * time.Second }

func (c *Client) Ping(send func(string)) {
	send(fmt.Sprintf(`{"ping":%d}`, client.Now()))
}

type ticker struct {
	Data struct {
		Symbol    string         `json:"symbol"`
		LastPrice jsonutil.Float `json:"lastPrice"`
		Volume24h jsonutil.Float `json:"volume24h"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "tickers.") || !strings.Contains(message, "lastPrice") {
		return nil
	}
	var t ticker
	if err := json.Unmarshal([]byte(message), &t); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(t.Data.Symbol)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: t.Data.LastPrice.V(), H24Volume: t.Data.Volume24h.V() * t.Data.LastPrice.V(),
	}}
}
