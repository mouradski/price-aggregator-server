package jsonutil

import (
	"strconv"
	"strings"
)

// Float is a float64 that unmarshals from either a JSON number or a JSON string
// (including an empty string, which decodes to 0). Crypto APIs are inconsistent
// about quoting numbers, so this absorbs both forms.
type Float float64

func (f *Float) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "" || s == "null" {
		*f = 0
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = Float(v)
	return nil
}

// V returns the underlying float64.
func (f Float) V() float64 { return float64(f) }
