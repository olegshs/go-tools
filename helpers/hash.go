package helpers

import (
	"crypto"
	"strings"
)

func HashByName(name string) crypto.Hash {
	var h crypto.Hash

	switch strings.ToLower(name) {
	case "md5":
		h = crypto.MD5
	case "sha1":
		h = crypto.SHA1
	case "sha224":
		h = crypto.SHA224
	case "sha256":
		h = crypto.SHA256
	case "sha384":
		h = crypto.SHA384
	case "sha512":
		h = crypto.SHA512
	}

	return h
}
