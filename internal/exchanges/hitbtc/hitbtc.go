package hitbtc

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "hitbtc" }

func (c *Client) URI(*client.Context) string { return "wss://api.hitbtc.com/api/3/ws/public" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			pairs = append(pairs, `"`+base+quote+`"`)
		}
	}
	send(fmt.Sprintf(`{"method":"subscribe","ch":"ticker/price/1s","params":{"symbols":[%s],"limit":1},"id":%d}`,
		strings.Join(pairs, ","), client.Now()))
}

type response struct {
	Data map[string]struct {
		C jsonutil.Float `json:"c"`
		Q jsonutil.Float `json:"q"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker/price/1s") {
		return nil
	}
	var r response
	if err := json.Unmarshal([]byte(message), &r); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for sym, d := range r.Data {
		base, quote := symbol.GetPair(sym)
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.C.V(), H24Volume: d.Q.V(),
		})
	}
	return out
}
