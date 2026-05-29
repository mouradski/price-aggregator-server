package model

type Source string

const (
	SourceWS   Source = "WS"
	SourceREST Source = "REST"
)

// VolumeUnavailable is the H24Volume sentinel for exchanges whose ticker does
// not provide (and from which we cannot compute) a 24h volume. It is
// distinct from a real 0, which means the volume is known to be zero.
const VolumeUnavailable = -1

type Ticker struct {
	LastPrice float64 `json:"lastPrice"`
	Exchange  string  `json:"exchange"`
	Base      string  `json:"base"`
	Quote     string  `json:"quote"`
	Timestamp int64   `json:"timestamp"`
	Source    Source  `json:"source"`
	H24Volume float64 `json:"h24Volume"`
}
