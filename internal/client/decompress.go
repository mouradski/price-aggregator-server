package client

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"
)

// decompress mirrors the binary onMessage handling of AbstractClientEndpoint:
// gzip is detected by magic bytes, otherwise zlib is attempted, finally falling
// back to treating the bytes as plain UTF-8 (some exchanges, e.g. Upbit, send
// uncompressed JSON inside binary frames). Raw DEFLATE is intentionally not
// attempted because flate.NewReader can return spurious bytes for plain JSON,
// which would mask the UTF-8 fallback.
func decompress(data []byte) string {
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		if s, ok := readAll(func() (io.ReadCloser, error) { return gzip.NewReader(bytes.NewReader(data)) }); ok {
			return s
		}
	}

	if s, ok := readAll(func() (io.ReadCloser, error) { return zlib.NewReader(bytes.NewReader(data)) }); ok {
		return s
	}

	return string(data)
}

func readAll(open func() (io.ReadCloser, error)) (string, bool) {
	r, err := open()
	if err != nil {
		return "", false
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil || len(out) == 0 {
		return "", false
	}
	return string(out), true
}
