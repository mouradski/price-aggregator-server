package deepcoin

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const url = "https://api.deepcoin.com/deepcoin/market/tickers?instType=SPOT"

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "deepcoin" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

type response struct {
	Data []struct {
		InstType string         `json:"instType"`
		InstID   string         `json:"instId"`
		Last     jsonutil.Float `json:"last"`
		AskPx    jsonutil.Float `json:"askPx"`
		BidPx    jsonutil.Float `json:"bidPx"`
		Vol24h   jsonutil.Float `json:"vol24h"`
	} `json:"data"`
}

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	var resp response
	if err := client.GetJSON(url, &resp); err != nil {
		return err
	}
	ts := client.Now()
	for _, d := range resp.Data {
		base, quote := symbol.GetPair(strings.ReplaceAll(d.InstID, "-SWAP", ""))
		if !ctx.IsAsset(base) || !ctx.IsQuote(quote) {
			continue
		}
		exch := c.Name()
		if d.InstType == "SWAP" {
			exch += "swap"
		}
		volume := d.Vol24h.V() * d.Last.V()
		push(model.Ticker{
			Source: model.SourceREST, Exchange: exch, Base: base, Quote: quote,
			LastPrice: d.Last.V(), Timestamp: ts, H24Volume: volume,
		})
		push(model.Ticker{
			Source: model.SourceREST, Exchange: exch + "-ask", Base: base, Quote: quote,
			LastPrice: (d.AskPx.V() + d.BidPx.V()) / 2, Timestamp: ts, H24Volume: volume,
		})
	}
	return nil
}
