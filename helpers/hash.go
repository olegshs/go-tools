package helpers

import (
	"crypto"
	"strings"
)

func HashByName(name string) crypto.Hash {
	var h crypto.Hash

	switch strings.ToLower(name) {
	case "md4":
		h = crypto.MD4
	case "md5":
		h = crypto.MD5
	case "sha-1", "sha1":
		h = crypto.SHA1
	case "sha-224", "sha224":
		h = crypto.SHA224
	case "sha-256", "sha256":
		h = crypto.SHA256
	case "sha-384", "sha384":
		h = crypto.SHA384
	case "sha-512", "sha512":
		h = crypto.SHA512
	case "sha-512/224":
		h = crypto.SHA512_224
	case "sha-512/256":
		h = crypto.SHA512_256
	case "sha3-224":
		h = crypto.SHA3_224
	case "sha3-256":
		h = crypto.SHA3_256
	case "sha3-384":
		h = crypto.SHA3_384
	case "sha3-512":
		h = crypto.SHA3_512
	case "md5+sha1":
		h = crypto.MD5SHA1
	case "ripemd-160":
		h = crypto.RIPEMD160
	case "blake2s-256":
		h = crypto.BLAKE2s_256
	case "blake2b-256":
		h = crypto.BLAKE2b_256
	case "blake2b-384":
		h = crypto.BLAKE2b_384
	case "blake2b-512":
		h = crypto.BLAKE2b_512
	}

	return h
}
