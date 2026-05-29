package crypto

import (
	"encoding/json"
	"fmt"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{ id int }

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "crypto" }

func (c *Client) URI(*client.Context) string { return "wss://stream.crypto.com/v2/market" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			c.id++
			send(fmt.Sprintf(`{"id":%d,"method":"subscribe","params":{"channels":["ticker.%s_%s"]},"nonce":%d}`,
				c.id, base, quote, client.Now()))
			c.id++
			send(fmt.Sprintf(`{"id":%d,"method":"subscribe","params":{"channels":["ticker.%s%s-PERP"]},"nonce":%d}`,
				c.id, base, quote, client.Now()))
		}
	}
}

// Pong answers crypto.com's heartbeat.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "public/heartbeat") {
		send(strings.ReplaceAll(message, "public/heartbeat", "public/respond-heartbeat"))
		return true
	}
	return false
}

type response struct {
	Result struct {
		InstrumentName string `json:"instrument_name"`
		Subscription   string `json:"subscription"`
		Data           []struct {
			A jsonutil.Float `json:"a"` // last price
			B jsonutil.Float `json:"b"` // best bid
			K jsonutil.Float `json:"k"` // best ask
			V jsonutil.Float `json:"v"` // volume
		} `json:"data"`
	} `json:"result"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker.") {
		return nil
	}
	var r response
	if err := json.Unmarshal([]byte(message), &r); err != nil || len(r.Result.Data) == 0 {
		return nil
	}

	future := strings.Contains(r.Result.InstrumentName, "-PERP")
	pairStr := strings.ReplaceAll(strings.ReplaceAll(r.Result.Subscription, "ticker.", ""), "-PERP", "")
	base, quote := symbol.GetPair(pairStr)
	d := r.Result.Data[0]
	volume := d.V.V() * d.A.V()

	name := c.Name()
	if future {
		name += "future"
	}
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: name, Base: base, Quote: quote,
			LastPrice: d.A.V(), H24Volume: volume},
		{Source: model.SourceWS, Timestamp: ts, Exchange: name + "-ask", Base: base, Quote: quote,
			LastPrice: (d.K.V() + d.B.V()) / 2, H24Volume: volume},
	}
}
