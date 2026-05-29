package toobit

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

func (c *Client) Name() string { return "toobit" }

func (c *Client) URI(*client.Context) string { return "wss://stream.toobit.com/quote/ws/v1" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.AssetsUpper() {
		pairs = append(pairs, base+"USDT")
	}
	send(fmt.Sprintf(`{"symbol":"%s","topic":"realtimes","event":"sub","params":{"binary":false}}`, strings.Join(pairs, ",")))
}

func (c *Client) PingInterval() time.Duration { return 60 * time.Second }

func (c *Client) Ping(send func(string)) {
	send(fmt.Sprintf(`{"ping":%d}`, client.Now()))
}

type tickerData struct {
	Symbol string `json:"symbol"`
	Data   []struct {
		C  jsonutil.Float `json:"c"`
		Qv jsonutil.Float `json:"qv"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "realtimes") {
		return nil
	}
	var t tickerData
	if err := json.Unmarshal([]byte(message), &t); err != nil || len(t.Data) == 0 {
		return nil
	}
	base, quote := symbol.GetPair(t.Symbol)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: t.Data[0].C.V(), H24Volume: t.Data[0].Qv.V(),
	}}
}
