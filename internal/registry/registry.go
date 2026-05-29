package registry

import (
	"ftso-prices/internal/client"
	"ftso-prices/internal/exchanges/ascendex"
	"ftso-prices/internal/exchanges/azbit"
	"ftso-prices/internal/exchanges/backpack"
	"ftso-prices/internal/exchanges/bequant"
	"ftso-prices/internal/exchanges/bigone"
	"ftso-prices/internal/exchanges/binance"
	"ftso-prices/internal/exchanges/binanceus"
	"ftso-prices/internal/exchanges/bingx"
	"ftso-prices/internal/exchanges/bitdelta"
	"ftso-prices/internal/exchanges/bitfinex"
	"ftso-prices/internal/exchanges/bitmart"
	"ftso-prices/internal/exchanges/bitmex"
	"ftso-prices/internal/exchanges/bitopro"
	"ftso-prices/internal/exchanges/bitrue"
	"ftso-prices/internal/exchanges/bitstamp"
	"ftso-prices/internal/exchanges/bitunix"
	"ftso-prices/internal/exchanges/bitvavo"
	"ftso-prices/internal/exchanges/bitvenus"
	"ftso-prices/internal/exchanges/blofin"
	"ftso-prices/internal/exchanges/bybit"
	"ftso-prices/internal/exchanges/cexio"
	"ftso-prices/internal/exchanges/coinbase"
	"ftso-prices/internal/exchanges/coinex"
	"ftso-prices/internal/exchanges/coinstore"
	"ftso-prices/internal/exchanges/cointr"
	"ftso-prices/internal/exchanges/coinw"
	"ftso-prices/internal/exchanges/crypto"
	"ftso-prices/internal/exchanges/deepcoin"
	"ftso-prices/internal/exchanges/deribit"
	"ftso-prices/internal/exchanges/digifinex"
	"ftso-prices/internal/exchanges/exmo"
	"ftso-prices/internal/exchanges/famex"
	"ftso-prices/internal/exchanges/fmfw"
	"ftso-prices/internal/exchanges/gateio"
	"ftso-prices/internal/exchanges/gemini"
	"ftso-prices/internal/exchanges/hashkey"
	"ftso-prices/internal/exchanges/hitbtc"
	"ftso-prices/internal/exchanges/huobi"
	"ftso-prices/internal/exchanges/hyperliquid"
	"ftso-prices/internal/exchanges/indoex"
	"ftso-prices/internal/exchanges/koinbay"
	"ftso-prices/internal/exchanges/kraken"
	"ftso-prices/internal/exchanges/kucoin"
	"ftso-prices/internal/exchanges/lbank"
	"ftso-prices/internal/exchanges/luno"
	"ftso-prices/internal/exchanges/mexc"
	"ftso-prices/internal/exchanges/nami"
	"ftso-prices/internal/exchanges/nonkyc"
	"ftso-prices/internal/exchanges/okex"
	"ftso-prices/internal/exchanges/orangex"
	"ftso-prices/internal/exchanges/ourbit"
	"ftso-prices/internal/exchanges/phemex"
	"ftso-prices/internal/exchanges/pionex"
	"ftso-prices/internal/exchanges/poloniex"
	"ftso-prices/internal/exchanges/toobit"
	"ftso-prices/internal/exchanges/weex"
	"ftso-prices/internal/exchanges/whitebit"
	"ftso-prices/internal/exchanges/xt"

	"ftso-prices/internal/exchanges/bit2me"
	"ftso-prices/internal/exchanges/btcturk"
	"ftso-prices/internal/exchanges/p2b"

	"ftso-prices/internal/exchanges/bibox"
	"ftso-prices/internal/exchanges/hotcoin"
	"ftso-prices/internal/exchanges/icrypex"

	"ftso-prices/internal/exchanges/biconomy"
	"ftso-prices/internal/exchanges/bullish"
	"ftso-prices/internal/exchanges/cryptomus"
	"ftso-prices/internal/exchanges/gleec"
	"ftso-prices/internal/exchanges/latoken"
	"ftso-prices/internal/exchanges/trubit"

	"ftso-prices/internal/exchanges/batonex"
	"ftso-prices/internal/exchanges/bitpanda"
	"ftso-prices/internal/exchanges/bitso"
	"ftso-prices/internal/exchanges/bluebit"
	"ftso-prices/internal/exchanges/btse"
	"ftso-prices/internal/exchanges/bydfi"
	"ftso-prices/internal/exchanges/cpatex"
	"ftso-prices/internal/exchanges/jucoin"
	"ftso-prices/internal/exchanges/pointpay"
	"ftso-prices/internal/exchanges/uzx"
	"ftso-prices/internal/exchanges/websea"

	"ftso-prices/internal/exchanges/bitget"
	"ftso-prices/internal/exchanges/upbit"
)

// allWs lists every websocket-based exchange implementation.
func allWs() []client.WsExchange {
	return []client.WsExchange{
		binance.New(),
		huobi.New(),
		kraken.New(),
		crypto.New(),
		bitfinex.New(),
		fmfw.New(),
		hitbtc.New(),
		exmo.New(),
		coinstore.New(),
		weex.New(),
		koinbay.New(),
		bingx.New(),
		ascendex.New(),
		bybit.New(),
		poloniex.New(),
		toobit.New(),
		cointr.New(),
		bequant.New(),
		hashkey.New(),
		bitvenus.New(),
		upbit.New(),
		cexio.New(),
		binanceus.New(),
		backpack.New(),
		hyperliquid.New(),
		deribit.New(),
		bitmex.New(),
	}
}

// allRest lists every REST-polling exchange implementation.
func allRest() []client.RestExchange {
	return []client.RestExchange{
		bitstamp.New(),
		deepcoin.New(),
		mexc.New(),
		gemini.New(),
		xt.New(),
		lbank.New(),
		phemex.New(),
		pionex.New(),
		bitvavo.New(),
		luno.New(),
		bigone.New(),
		nonkyc.New(),
		bitopro.New(),
		nami.New(),
		indoex.New(),
		famex.New(),
		azbit.New(),
		coinw.New(),
		bitrue.New(),
		bitunix.New(),
		bitdelta.New(),
		btcturk.New(),
		p2b.New(),
		bit2me.New(),
		bibox.New(),
		hotcoin.New(),
		icrypex.New(),
		latoken.New(),
		trubit.New(),
		cryptomus.New(),
		gleec.New(),
		bullish.New(),
		biconomy.New(),
		batonex.New(),
		bitso.New(),
		bitpanda.New(),
		btse.New(),
		websea.New(),
		jucoin.New(),
		bydfi.New(),
		cpatex.New(),
		pointpay.New(),
		uzx.New(),
		bluebit.New(),
		bitget.New(),
		ourbit.New(),
	}
}

// Hybrid pairs a websocket primary feed with a REST seed/fallback for the same
// exchange.
type Hybrid struct {
	WS   client.WsExchange
	REST client.RestExchange
}

// allHybrid lists exchanges that have both a websocket feed (primary) and a
// REST feed (startup seed + fallback). The WS and REST implementations share
// the same exchange name.
func allHybrid() []Hybrid {
	return []Hybrid{
		{okex.NewWS(), okex.New()},
		{coinbase.NewWS(), coinbase.New()},
		{gateio.NewWS(), gateio.New()},
		{whitebit.NewWS(), whitebit.New()},
		{coinex.NewWS(), coinex.New()},
		{bitmart.NewWS(), bitmart.New()},
		{digifinex.NewWS(), digifinex.New()},
		{kucoin.NewWS(), kucoin.New()},
		{blofin.New(), blofin.NewRest()},
		{orangex.New(), orangex.NewRest()},
	}
}

// EnabledHybrid returns the hybrid exchanges permitted by the config filter.
func EnabledHybrid(enabled func(string) bool) []Hybrid {
	var out []Hybrid
	for _, h := range allHybrid() {
		if enabled(h.WS.Name()) {
			out = append(out, h)
		}
	}
	return out
}

// EnabledWs returns the websocket exchanges permitted by the config filter.
func EnabledWs(enabled func(string) bool) []client.WsExchange {
	var out []client.WsExchange
	for _, ex := range allWs() {
		if enabled(ex.Name()) {
			out = append(out, ex)
		}
	}
	return out
}

// EnabledRest returns the REST exchanges permitted by the config filter.
func EnabledRest(enabled func(string) bool) []client.RestExchange {
	var out []client.RestExchange
	for _, ex := range allRest() {
		if enabled(ex.Name()) {
			out = append(out, ex)
		}
	}
	return out
}
