package kucoin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ftso-prices/internal/client"
	"ftso-prices/internal/jsonutil"
	"ftso-prices/internal/model"
	"ftso-prices/internal/symbol"
)

const bulletURL = "https://api.kucoin.com/api/v1/bullet-public"

// WSClient streams KuCoin tickers over websocket. KuCoin requires a short-lived
// token fetched over REST before connecting; the /market/ticker:all topic
// streams every symbol's last price + best bid/ask (but no 24h volume, which
// the REST fallback still provides). Shares the "kucoin" name with the REST
// Client.
type WSClient struct{ id int }

func NewWS() *WSClient { return &WSClient{} }

func (c *WSClient) Name() string { return "kucoin" }

// URI fetches a fresh bullet token and returns the tokenised websocket URL. It
// runs on every (re)connect, so reconnects always use a valid token.
func (c *WSClient) URI(*client.Context) string {
	resp, err := http.Post(bulletURL, "application/json", nil)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	var r struct {
		Data struct {
			Token   string `json:"token"`
			Servers []struct {
				Endpoint string `json:"endpoint"`
			} `json:"instanceServers"`
		} `json:"data"`
	}
	if json.Unmarshal(body, &r) != nil || len(r.Data.Servers) == 0 || r.Data.Token == "" {
		return ""
	}
	return r.Data.Servers[0].Endpoint + "?token=" + r.Data.Token + "&connectId=ftso"
}

func (c *WSClient) Subscribe(_ *client.Context, send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"id":%d,"type":"subscribe","topic":"/market/ticker:all","response":true}`, c.id))
}

func (c *WSClient) PingInterval() time.Duration { return 15 * time.Second }

func (c *WSClient) Ping(send func(string)) {
	c.id++
	send(fmt.Sprintf(`{"id":"%d","type":"ping"}`, c.id))
}

type wsMessage struct {
	Topic   string `json:"topic"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Data    struct {
		Price   jsonutil.Float `json:"price"`
		BestBid jsonutil.Float `json:"bestBid"`
		BestAsk jsonutil.Float `json:"bestAsk"`
	} `json:"data"`
}

func (c *WSClient) MapTicker(ctx *client.Context, message string) []model.Ticker {
	if !strings.Contains(message, "ticker:all") || !strings.Contains(message, "price") {
		return nil
	}
	var m wsMessage
	if err := json.Unmarshal([]byte(message), &m); err != nil || m.Subject == "" {
		return nil
	}
	base, quote := symbol.GetPair(m.Subject)
	ts := client.Now()
	return []model.Ticker{
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name(), Base: base, Quote: quote,
			LastPrice: m.Data.Price.V(), H24Volume: model.VolumeUnavailable},
		{Source: model.SourceWS, Timestamp: ts, Exchange: c.Name() + "-ask", Base: base, Quote: quote,
			LastPrice: (m.Data.BestBid.V() + m.Data.BestAsk.V()) / 2, H24Volume: model.VolumeUnavailable},
	}
}
