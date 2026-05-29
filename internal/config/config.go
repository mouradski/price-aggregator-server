package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// AllQuotes mirrors Constants.ALL_QUOTES from the Java implementation.
var AllQuotes = []string{"usd", "usdt", "usdc", "usds", "fdusd", "tusd", "usdd", "usde", "dai"}

// defaultAssets mirrors Constants.SYMBOLS, used when ASSETS is not set.
var defaultAssets = []string{
	"xrp", "btc", "eth", "algo", "xlm", "ada", "matic", "sol", "fil", "flr",
	"sgb", "doge", "xdc", "arb", "avax", "bnb", "usdc", "busd", "usdt",
}

type Config struct {
	Assets   []string        // lowercase
	AssetSet map[string]bool // lowercase membership
	Quotes   []string        // lowercase
	Port     int
	Timeout  time.Duration
	// Whitelist/blacklist of exchange names. Empty whitelist means all enabled.
	Whitelist map[string]bool
	Blacklist map[string]bool
}

// Load reads configuration from the environment, optionally seeded from a .env
// file found at envPath (ignored if it does not exist).
func Load(envPath string) *Config {
	loadDotEnv(envPath)

	assets := splitCSV(get("ASSETS", ""))
	if len(assets) == 0 {
		assets = defaultAssets
	}
	assetSet := make(map[string]bool, len(assets))
	for _, a := range assets {
		assetSet[strings.ToLower(a)] = true
	}

	port := 8090
	if v := get("AGGREGATOR_PORT", get("QUARKUS_HTTP_PORT", "")); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	timeout := 30 * time.Second
	if v := get("MESSAGE_TIMEOUT", get("DEFAULT_MESSAGE_TIMEOUT", "")); v != "" {
		if t, err := strconv.Atoi(v); err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}

	return &Config{
		Assets:    assets,
		AssetSet:  assetSet,
		Quotes:    AllQuotes,
		Port:      port,
		Timeout:   timeout,
		Whitelist: toSet(splitCSV(get("EXCHANGES", get("EXCHANGE", "")))),
		Blacklist: toSet(splitCSV(get("EXCHANGES_EXCLUDED", get("EXCHANGE_EXCLUDED", "")))),
	}
}

// Enabled applies the same precedence as AbstractClientEndpoint.isExchangeEnabled:
// blacklist wins, then whitelist (with "all"), otherwise enabled by default.
func (c *Config) Enabled(name string) bool {
	if c.Blacklist[name] {
		return false
	}
	if len(c.Whitelist) > 0 {
		return c.Whitelist["all"] || c.Whitelist[name]
	}
	return true
}

// AssetsUpper returns the assets uppercased.
func (c *Config) AssetsUpper() []string { return upper(c.Assets) }

// QuotesUpper returns the quotes uppercased.
func (c *Config) QuotesUpper() []string { return upper(c.Quotes) }

func upper(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}

func get(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func toSet(items []string) map[string]bool {
	m := make(map[string]bool, len(items))
	for _, it := range items {
		m[it] = true
	}
	return m
}

// loadDotEnv populates env vars from a .env file without overriding existing ones.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, val)
		}
	}
}
