package encoder

import (
	"encoding/base64"
)

// Base64 encodes data as standard base64, for text columns.
type Base64 struct {
}

// Encode returns src encoded as base64.
func (encoder *Base64) Encode(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))

	base64.StdEncoding.Encode(dst, src)

	return dst, nil
}

// Decode returns src decoded from base64.
func (encoder *Base64) Decode(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))

	n, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	// DecodedLen is the maximum length; padding makes the result shorter.
	return dst[:n], nil
}
