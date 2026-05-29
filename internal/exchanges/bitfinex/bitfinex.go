package bitfinex

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type pair struct{ base, quote string }

type Client struct {
	channels map[float64]pair
}

func New() *Client { return &Client{channels: make(map[float64]pair)} }

func (c *Client) Name() string { return "bitfinex" }

func (c *Client) URI(*client.Context) string { return "wss://api-pub.bitfinex.com/ws/2" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			send(fmt.Sprintf(`{"event":"subscribe", "channel":"ticker","symbol":"t%s%s"}`, base, bitfinexQuote(quote)))
		}
	}
}

func bitfinexQuote(quote string) string {
	switch quote {
	case "USDT":
		return "UST"
	case "USDC":
		return "UDC"
	default:
		return quote
	}
}

// DecodeMetadata records the channelId -> pair mapping from subscription acks.
func (c *Client) DecodeMetadata(message string, _ func(string)) {
	if !strings.Contains(message, "subscribed") {
		return
	}
	var ack struct {
		ChanID float64 `json:"chanId"`
		Symbol string  `json:"symbol"`
	}
	if err := json.Unmarshal([]byte(message), &ack); err != nil || ack.Symbol == "" {
		return
	}
	symbolID := strings.NewReplacer("UST", "USDT", "UDC", "USDC", ":", "").Replace(strings.TrimPrefix(ack.Symbol, "t"))
	base, quote := symbol.GetPair(symbolID)
	c.channels[ack.ChanID] = pair{base, quote}
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if strings.Contains(message, "hb") || strings.Contains(message, "event") {
		return nil
	}

	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(message), &arr); err != nil || len(arr) < 2 {
		return nil
	}
	var channelID float64
	if err := json.Unmarshal(arr[0], &channelID); err != nil {
		return nil
	}
	p, ok := c.channels[channelID]
	if !ok {
		return nil
	}

	var td []float64
	if err := json.Unmarshal(arr[1], &td); err != nil || len(td) < 8 {
		return nil
	}
	last := td[6]
	ask := (td[0] + td[2]) / 2
	volume := td[7] * last

	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: p.base, Quote: p.quote,
			LastPrice: last, H24Volume: volume},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: p.base, Quote: p.quote,
			LastPrice: ask, H24Volume: volume},
	}
}
