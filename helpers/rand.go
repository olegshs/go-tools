package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

func RandBytes(length int) []byte {
	if length <= 0 {
		return []byte{}
	}

	b := make([]byte, length)

	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}

func RandHex(length int) string {
	if length <= 0 {
		return ""
	}

	bytesLength := (length-1)/2 + 1

	b := RandBytes(bytesLength)

	str := hex.EncodeToString(b)
	str = str[:length]

	return str
}
