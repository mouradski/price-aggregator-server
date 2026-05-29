package jsonutil

import (
	"bytes"
	"encoding/json"
)

// String is a string that tolerates a JSON value which is sometimes a number or
// null instead of a string (some exchanges, e.g. Nami, emit 0 for absent
// symbol/currency fields). Non-string values decode to "".
type String string

func (s *String) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || b[0] != '"' {
		*s = ""
		return nil
	}
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = String(str)
	return nil
}

// V returns the underlying string.
func (s String) V() string { return string(s) }
