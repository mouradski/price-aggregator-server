package jupiter

import (
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
)

// Jupiter is a Solana on-chain price aggregator. It prices SPL tokens by mint
// address and returns a USD-denominated price (no bid/ask, no 24h volume), so
// we keep a curated map of the Solana-native assets we track. Mint addresses
// verified via https://lite-api.jup.ag/tokens/v2/search.
const priceURL = "https://lite-api.jup.ag/price/v3?ids="

var mints = map[string]string{
	"sol":   "So11111111111111111111111111111111111111112",
	"bonk":  "DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263",
	"wif":   "EKpQGSJtjMFqKZ9KQanSqYXRcF8fBopzLHYxdM65zcjm",
	"jup":   "JUPyiwrYJFskUPiHa7hkeR8VUtAeFoSYbKedZNsDvCN",
	"pump":  "pumpCmXqMfrsAkQ5r49WcJnRayYRqmXz6ae8H7H9Dfn",
	"pengu": "2zMMhcVQEXDtdE6vsFS7S7D5oUodfJHE8vd1gnBouauv",
	"trump": "6p6xgHyF7AeE6TZkSmFsko444wqoP15icUSqi2jfGiPN",
}

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "jupiter" }

func (c *Client) Interval() time.Duration { return 3 * time.Second }

func (c *Client) Poll(ctx *client.Context, push func(model.Ticker)) error {
	// Build the request only for assets we are configured to track, and keep a
	// reverse mint -> asset map to decode the response.
	var ids []string
	byMint := make(map[string]string)
	for asset, mint := range mints {
		if ctx.IsAsset(asset) {
			ids = append(ids, mint)
			byMint[mint] = asset
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var prices map[string]struct {
		UsdPrice jsonutil.Float `json:"usdPrice"`
	}
	if err := client.GetJSON(priceURL+strings.Join(ids, ","), &prices); err != nil {
		return err
	}

	ts := client.Now()
	for mint, p := range prices {
		asset, ok := byMint[mint]
		if !ok || p.UsdPrice.V() == 0 {
			continue
		}
		push(model.Ticker{Source: model.SourceREST, Exchange: c.Name(),
			Base: strings.ToUpper(asset), Quote: "USD",
			LastPrice: p.UsdPrice.V(), Timestamp: ts, H24Volume: model.VolumeUnavailable})
	}
	return nil
}
