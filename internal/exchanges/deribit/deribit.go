package deribit

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// Deribit perpetuals are USD-denominated.
const quote = "USD"

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "deribit" }

func (c *Client) URI(*client.Context) string { return "wss://www.deribit.com/ws/api/v2" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	if !ctx.IsQuote(quote) {
		return
	}
	// One subscribe per instrument so a non-existent perpetual can't break the batch.
	for _, base := range ctx.AssetsUpper() {
		c.id++
		send(fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"method":"public/subscribe","params":{"channels":["ticker.%s-PERPETUAL.100ms"]}}`,
			c.id, base))
	}
}

type wsMessage struct {
	Params struct {
		Channel string `json:"channel"`
		Data    struct {
			InstrumentName string         `json:"instrument_name"`
			LastPrice      jsonutil.Float `json:"last_price"`
			BestBidPrice   jsonutil.Float `json:"best_bid_price"`
			BestAskPrice   jsonutil.Float `json:"best_ask_price"`
			Stats          struct {
				VolumeUSD jsonutil.Float `json:"volume_usd"`
			} `json:"stats"`
		} `json:"data"`
	} `json:"params"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"channel":"ticker.`) || !strings.Contains(message, "last_price") {
		return nil
	}
	var m wsMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	base := strings.TrimSuffix(m.Params.Data.InstrumentName, "-PERPETUAL")
	if base == "" || !ctx.IsAsset(base) {
		return nil
	}
	d := m.Params.Data
	vol := d.Stats.VolumeUSD.V()
	ts := client.Now()
	out := []model.Ticker{{
		Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: d.LastPrice.V(), H24Volume: vol,
	}}
	if d.BestBidPrice.V() != 0 && d.BestAskPrice.V() != 0 {
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (d.BestBidPrice.V() + d.BestAskPrice.V()) / 2, H24Volume: vol,
		})
	}
	return out
}
