package backpack

import (
	"encoding/json"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Backpack lists USDC/USDT spot quotes.
var wsQuotes = []string{"USDC", "USDT"}

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "backpack" }

func (c *Client) URI(*client.Context) string { return "wss://ws.backpack.exchange" }

func (c *Client) Subscribe(ctx *client.Context, send func(string)) {
	var params []string
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range wsQuotes {
			if base == quote || !ctx.IsQuote(quote) {
				continue
			}
			params = append(params, `"ticker.`+base+"_"+quote+`"`)
		}
	}
	send(`{"method":"SUBSCRIBE","params":[` + strings.Join(params, ",") + `]}`)
}

// Go's JSON matching is case-insensitive, and the payload has key pairs that
// differ only by case: "e"(string event)/"E"(numeric time) and "v"(base
// vol)/"V"(quote vol). Each case-variant needs its own exact-tagged field, else
// e.g. the string "ticker" from "e" lands on a numeric field and fails the whole
// unmarshal.
type wsMessage struct {
	Data struct {
		Symbol      string          `json:"s"`
		LastPrice   jsonutil.Float  `json:"c"`
		QuoteVolume jsonutil.Float  `json:"V"`
		Event       jsonutil.String `json:"e"`
		EventTime   jsonutil.Float  `json:"E"`
		BaseVolume  jsonutil.Float  `json:"v"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, `"ticker`) || !strings.Contains(message, `"c"`) {
		return nil
	}
	var m wsMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil || m.Data.Symbol == "" {
		return nil
	}
	base, quote := symbol.GetPair(m.Data.Symbol)
	return []model.Ticker{{
		Source: model.SourceWS, Timestamp: client.Now(), Exchange: c.Name(), Base: base, Quote: quote,
		LastPrice: m.Data.LastPrice.V(), H24Volume: m.Data.QuoteVolume.V(),
	}}
}
