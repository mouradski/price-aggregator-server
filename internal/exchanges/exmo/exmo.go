package exmo

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Exmo only quotes against USDT and USDC (getAllStablecoinQuotesExceptBusd).
var stablecoinQuotes = []string{"USDT", "USDC"}

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "exmo" }

func (c *Client) URI(*client.Context) string { return "wss://ws-api.exmo.com:443/v1/public" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range stablecoinQuotes {
			c.id++
			send(fmt.Sprintf(`{"id":%d,"method":"subscribe","topics":["spot/ticker:%s_%s"]}`, c.id, base, quote))
		}
	}
}

type update struct {
	Topic string `json:"topic"`
	Data  struct {
		BuyPrice  jsonutil.Float `json:"buy_price"`
		SellPrice jsonutil.Float `json:"sell_price"`
		LastTrade jsonutil.Float `json:"last_trade"`
		VolCurr   jsonutil.Float `json:"vol_curr"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "spot/ticker:") {
		return nil
	}
	var u update
	if err := json.Unmarshal([]byte(message), &u); err != nil {
		return nil
	}
	base, quote := symbol.GetPair(strings.ReplaceAll(u.Topic, "spot/ticker:", ""))
	vol := u.Data.VolCurr.V()
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: u.Data.LastTrade.V(), H24Volume: vol},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (u.Data.BuyPrice.V() + u.Data.SellPrice.V()) / 2, H24Volume: vol},
	}
}
