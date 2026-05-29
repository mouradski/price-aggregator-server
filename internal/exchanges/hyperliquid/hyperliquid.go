package hyperliquid

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
)

// Hyperliquid perps are USDC-collateralised; mids are USD-denominated prices.
const quote = "USDC"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "hyperliquid" }

func (c *Client) URI(*client.Context) string { return "wss://api.hyperliquid.xyz/ws" }

func (c *Client) Subscribe(_ *client.Context, send func(string)) {
	send(`{"method":"subscribe","subscription":{"type":"allMids"}}`)
}

func (c *Client) PingInterval() time.Duration { return 50 * time.Second }

func (c *Client) Ping(send func(string)) { send(`{"method":"ping"}`) }

type wsMessage struct {
	Channel string `json:"channel"`
	Data    struct {
		Mids map[string]string `json:"mids"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "allMids") {
		return nil
	}
	var m wsMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	if !ctx.IsQuote(quote) {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for coin, px := range m.Data.Mids {
		// Skip spot-index keys (e.g. "#1000", "@1", "kPEPE"); keep configured assets.
		if !ctx.IsAsset(coin) {
			continue
		}
		last, err := strconv.ParseFloat(px, 64)
		if err != nil {
			continue
		}
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(),
			Base: strings.ToUpper(coin), Quote: quote, LastPrice: last,
			H24Volume: model.VolumeUnavailable,
		})
	}
	return out
}
