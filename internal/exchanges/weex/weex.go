package weex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

// Weex only streams a fixed set of symbols.
var symbols = []string{
	"BTCUSDT", "BTCUSDC", "ETHUSDT", "ETHUSDC", "SOLUSDT", "SOLUSDC", "XRPUSDT", "XRPUSDC",
	"BNBUSDT", "BNBUSDC", "DOGEUSDT", "DOGEUSDC", "ADAUSDT", "ADAUSDC", "LINKUSDT", "LINKUSDC",
	"LTCUSDT", "LTCUSDC", "SUIUSDT", "SUIUSDC", "AVAXUSDT", "AVAXUSDC", "SHIBUSDT", "ARBUSDT",
	"ARBUSDC", "TRUMPUSDT", "PEPEUSDT", "PEPEUSDC", "BCHUSDT", "BCHUSDC", "TRXUSDT", "TRXUSDC",
	"NOTUSDT", "NOTUSDC", "WIFUSDT", "WIFUSDC", "TONUSDT", "APTUSDT", "APTUSDC", "ETCUSDT",
	"AAVEUSDT", "AAVEUSDC", "RUNEUSDT", "RUNEUSDC", "ALGOUSDT", "ALGOUSDC", "NEARUSDT", "NEARUSDC",
	"ICPUSDC", "FILUSDT", "FILUSDC", "OPUSDT", "OPUSDC", "TAOUSDT", "JUPUSDC", "PYTHUSDT",
	"RENDERUSDC", "UNIUSDT", "UNIUSDC", "XLMUSDT", "MATICUSDT", "MATICUSDC", "HNTUSDT", "HBARUSDT",
	"BERAUSDT", "SUSDT", "SUSDC", "FTMUSDT", "ENAUSDC", "ETHFIUSDT", "ETHFIUSDC", "FETUSDC",
	"PAXGUSDT", "USDCUSDT",
}

type Client struct{}

func New() *Client { return &Client{} }

func (c *Client) Name() string { return "weex" }

func (c *Client) URI(*client.Context) string { return "wss://ws-spot.weex.com/v2/ws/public" }

func (c *Client) Headers() http.Header {
	return http.Header{
		"Origin":     {"https://www.weex.com"},
		"User-Agent": {"Mozilla/5.0 Chrome/124.0 Safari/537.36"},
	}
}

func (c *Client) Subscribe(_ *client.Context, send func(string)) {
	for _, s := range symbols {
		send(fmt.Sprintf(`{"event":"subscribe","channel":"ticker.%s_SPBL"}`, s))
	}
}

// Pong answers Weex's ping by echoing it back as a pong.
func (c *Client) Pong(message string, send func(string)) bool {
	if strings.Contains(message, "ping") {
		send(strings.ReplaceAll(message, "ping", "pong"))
		return true
	}
	return false
}

type event struct {
	Data []struct {
		Symbol    string         `json:"symbol"`
		LastPrice jsonutil.Float `json:"lastPrice"`
	} `json:"data"`
}

func (c *Client) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "symbol") {
		return nil
	}
	var e event
	if err := json.Unmarshal([]byte(message), &e); err != nil {
		return nil
	}
	ts := client.Now()
	var out []model.Ticker
	for _, d := range e.Data {
		base, quote := symbol.GetPair(strings.Split(d.Symbol, "_")[0])
		out = append(out, model.Ticker{
			Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: d.LastPrice.V(), H24Volume: model.VolumeUnavailable,
		})
	}
	return out
}
