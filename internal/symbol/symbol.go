package symbol

import "strings"

// knownQuotes are sorted longest-first so that e.g. "FDUSD" is matched before
// "USD", otherwise "BTCFDUSD" would wrongly match "USD".
var knownQuotes = []string{
	"FDUSD", "USDT", "USDC", "USDS", "USDD", "USDE", "TUSD", "BUSD", "USD", "DAI",
}

// stablecoinBases are stablecoins that may legitimately appear as the base
// currency of a pair (e.g. USDTBUSD, USDCUSDT).
var stablecoinBases = []string{
	"FDUSD", "USDT", "USDC", "USDS", "USDD", "USDE", "TUSD", "BUSD", "DAI",
}

// GetPair splits a remote pair string into (base, quote), both uppercase.
func GetPair(remotePair string) (string, string) {
	pair := clean(remotePair)
	base := baseCurrency(pair)
	quote := strings.TrimPrefix(pair, base)
	return base, quote
}

func clean(remotePair string) string {
	r := strings.NewReplacer("-", "", "_", "", "/", "", ":", "")
	return strings.ToUpper(r.Replace(remotePair))
}

func baseCurrency(pair string) string {
	// Case 1: base itself is a stablecoin (e.g. USDTBUSD, USDCUSDT).
	for _, sc := range stablecoinBases {
		if strings.HasPrefix(pair, sc) && len(pair) > len(sc) {
			remainder := pair[len(sc):]
			for _, q := range knownQuotes {
				if remainder == q {
					return sc
				}
			}
		}
	}

	// Case 2: known quote as suffix, longest first.
	for _, q := range knownQuotes {
		if strings.HasSuffix(pair, q) && len(pair) > len(q) {
			return pair[:len(pair)-len(q)]
		}
	}

	// Case 3: unrecognised quote, return pair as-is.
	return pair
}
