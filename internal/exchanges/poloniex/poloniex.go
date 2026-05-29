package poloniex

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

func (c *Client) Name() string { return "poloniex" }

func (c *Client) URI(*client.Context) string { return "wss://ws.poloniex.com/ws/public" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			pairs = append(pairs, `"`+base+"_"+quote+`"`)
		}
	}
	send(fmt.Sprintf(`{"event":"subscribe","channel":["ticker"],"symbols":[%s]}`, strings.Join(pairs, ",")))
}

func (c *Client) PingInterval() time.Duration { return 30 * time.Second }

func (c *Client) Ping(send func(string)) { send(`{"event":"ping"}`) }

type tickerMessage struct {
	Data []struct {
		Symbol string         `json:"symbol"`
		Close  jsonutil.Float `json:"close"`
		Amount jsonutil.Float `json:"amount"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker") {
		return nil
	}
	var m tickerMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range m.Data {
		base, quote := symbol.GetPair(d.Symbol)
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.Close.V(), H24Volume: d.Amount.V(),
		})
	}
	return out
}
