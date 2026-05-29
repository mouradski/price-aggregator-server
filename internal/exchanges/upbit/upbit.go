package upbit

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "upbit" }

func (c *Client) URI(*client.Context) string { return "wss://api.upbit.com/websocket/v1" }

// Upbit only offers USDT among the configured quote currencies; sending codes
// for non-existent markets makes Upbit stream nothing, so restrict to USDT.
var supportedQuotes = []string{"USDT"}

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var codes []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range supportedQuotes {
			if !ctx.IsQuote(quote) {
				continue
			}
			// Upbit market codes are formatted QUOTE-BASE (e.g. USDT-BTC).
			codes = append(codes, `"`+quote+"-"+base+`"`)
		}
	}
	send(`[{"ticket":"tickers"},{"type":"ticker","codes":[` + strings.Join(codes, ",") + `]}]`)
}

type ticker struct {
	Code          string         `json:"code"`
	TradePrice    jsonutil.Float `json:"trade_price"`
	AccTradePrice jsonutil.Float `json:"acc_trade_price_24h"` // 24h quote volume
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker") {
		return nil
	}
	var t ticker
	if err := json.Unmarshal([]byte(message), &t); err != nil {
		return nil
	}
	// Code is QUOTE-BASE; split directly rather than guessing the boundary.
	parts := strings.SplitN(t.Code, "-", 2)
	if len(parts) != 2 {
		return nil
	}
	quote, base := parts[0], parts[1]
	if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
		return nil
	}
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: t.TradePrice.V(), H24Volume: t.AccTradePrice.V(),
	}}
}
