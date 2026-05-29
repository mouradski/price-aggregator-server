package kraken

import (
	"encoding/json"
	"strconv"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "kraken" }

func (c *Client) URI(*client.Context) string { return "wss://ws.kraken.com" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var pairs []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			pairs = append(pairs, `"`+base+"/"+quote+`"`)
		}
	}
	send(`{"event":"subscribe", "pair":[` + strings.Join(pairs, ",") + `], "subscription":{"name":"ticker"}}`)
}

// The c/a/b/v arrays mix strings and numbers (e.g. ["73778.2",0,"0.0013"]),
// so each element is kept raw and parsed individually.
type tick struct {
	C []json.RawMessage `json:"c"`
	A []json.RawMessage `json:"a"`
	B []json.RawMessage `json:"b"`
	V []json.RawMessage `json:"v"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker") ||
		strings.Contains(message, "systemStatus") ||
		strings.Contains(message, "errorMessage") ||
		strings.Contains(message, "subscriptionStatus") {
		return nil
	}

	var arr []json.RawMessage
	if err := json.Unmarshal([]byte(message), &arr); err != nil || len(arr) < 4 {
		return nil
	}

	var details tick
	if err := json.Unmarshal(arr[1], &details); err != nil {
		return nil
	}
	var pairStr string
	if err := json.Unmarshal(arr[3], &pairStr); err != nil {
		return nil
	}
	if len(details.C) < 1 || len(details.A) < 1 || len(details.B) < 1 || len(details.V) < 2 {
		return nil
	}

	pairStr = strings.ReplaceAll(strings.ReplaceAll(pairStr, "XBT", "BTC"), "XDG", "DOGE")
	base, quote := symbol.GetPair(pairStr)

	last := parse(details.C[0])
	ask := parse(details.A[0])
	bid := parse(details.B[0])
	volume := last * parse(details.V[1])
	avg := (ask + bid) / 2

	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: last, H24Volume: volume},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: avg, H24Volume: volume},
	}
}

func parse(raw json.RawMessage) float64 {
	v, _ := strconv.ParseFloat(strings.Trim(string(raw), `"`), 64)
	return v
}
