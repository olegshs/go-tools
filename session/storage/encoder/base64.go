package encoder

import (
	"encoding/base64"
)

type Base64 struct {
}

func (encoder *Base64) Encode(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))

	base64.StdEncoding.Encode(dst, src)

	return dst, nil
}

func (encoder *Base64) Decode(src []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(src)))

	_, err := base64.StdEncoding.Decode(dst, src)
	if err != nil {
		return nil, err
	}

	return dst, nil
}
