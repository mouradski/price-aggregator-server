package gemini

import (
	"strconv"
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

const (
	symbolsURL = "https://api.gemini.com/v1/symbols"
	tickerURL  = "https://api.gemini.com/v1/pubticker/"
)

type Client struct {
	symbols map[string]bool
}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "gemini" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

func (c *Client) loadSymbols() {
	if c.symbols != nil {
		return
	}
	c.symbols = make(map[string]bool)
	var syms []string
	if err := client.GetJSON(symbolsURL, &syms); err != nil {
		return
	}
	for _, s := range syms {
		c.symbols[strings.ToUpper(s)] = true
	}
}

type ticker struct {
	Bid    jsonutil.Float         `json:"bid"`
	Ask    jsonutil.Float         `json:"ask"`
	Last   jsonutil.Float         `json:"last"`
	Volume map[string]interface{} `json:"volume"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	c.loadSymbols()
	if len(c.symbols) == 0 {
		return nil
	}
	ts := client.Now()
	for _, base := range ctx.AssetsUpper() {
		for _, quote := range ctx.QuotesUpper() {
			sym := base + quote
			if !c.symbols[sym] {
				continue
			}
			var t ticker
			if err := client.GetJSON(tickerURL+sym, &t); err != nil {
				continue
			}
			vol := volumeFor(t.Volume, quote)
			push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(), Base: base, Quote: quote,
				LastPrice: t.Last.V(), Timestamp: ts, H24Volume: vol})
			if t.Ask.V() != 0 && t.Bid.V() != 0 {
				push(model.Ticker{Source: model.SourceREST, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
					LastPrice: (t.Ask.V() + t.Bid.V()) / 2, Timestamp: ts, H24Volume: vol})
			}
		}
	}
	return nil
}

func volumeFor(m map[string]interface{}, quote string) float64 {
	v, ok := m[quote]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return n
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	default:
		return 0
	}
}
