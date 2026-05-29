package symbol

import "testing"

func TestGetPair(t *testing.T) {
	cases := []struct {
		in, base, quote string
	}{
		{"BTCUSDT", "BTC", "USDT"},
		{"btc-usd", "BTC", "USD"},
		{"ETH_USDC", "ETH", "USDC"},
		{"BTCFDUSD", "BTC", "FDUSD"},
		{"USDTBUSD", "USDT", "BUSD"},
		{"USDCUSDT", "USDC", "USDT"},
		{"XRP/DAI", "XRP", "DAI"},
		{"BTCETH", "BTCETH", ""}, // ETH is not a known quote
	}
	for _, c := range cases {
		base, quote := GetPair(c.in)
		if base != c.base || quote != c.quote {
			t.Errorf("GetPair(%q) = (%q,%q), want (%q,%q)", c.in, base, quote, c.base, c.quote)
		}
	}
}
