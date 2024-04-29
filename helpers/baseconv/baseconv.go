// Пакет baseconv предоставляет функции для преобразования данных
// при помощи различных алфавитов кодирования.
package baseconv

import (
	"fmt"
	"math/big"
)

const (
	AlphabetBase36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	AlphabetBase58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	AlphabetBase62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func Parse(src string, alphabet string) (*big.Int, error) {
	negative := src[0] == '-'
	if negative {
		src = src[1:]
	}

	dst := new(big.Int)
	radix := big.NewInt(int64(len(alphabet)))

	decodeMap := make([]byte, 256)
	for i, l := 0, len(decodeMap); i < l; i++ {
		decodeMap[i] = 0xFF
	}
	for i, l := 0, len(alphabet); i < l; i++ {
		decodeMap[alphabet[i]] = byte(i)
	}

	for i, l := 0, len(src); i < l; i++ {
		b := decodeMap[src[i]]
		if b == 0xFF {
			return nil, fmt.Errorf("invalid character at offset %d", i)
		}

		dst.Mul(dst, radix)
		dst.Add(dst, big.NewInt(int64(b)))
	}

	if negative {
		dst.Neg(dst)
	}

	return dst, nil
}

func FormatInt(src int64, alphabet string) string {
	n := new(big.Int)
	n.SetInt64(src)

	return format(n, alphabet)
}

func FormatBigInt(src *big.Int, alphabet string) string {
	n := new(big.Int)
	n.Set(src)

	return format(n, alphabet)
}

func FormatBytes(src []byte, alphabet string) string {
	n := new(big.Int)
	n.SetBytes(src)

	return format(n, alphabet)
}

func FormatDecimal(src string, alphabet string) string {
	n := new(big.Int)
	n.SetString(src, 10)

	return format(n, alphabet)
}

func FormatHex(src string, alphabet string) string {
	n := new(big.Int)
	n.SetString(src, 16)

	return format(n, alphabet)
}

func format(src *big.Int, alphabet string) string {
	zero := big.NewInt(0)
	cmpZero := src.Cmp(zero)

	if cmpZero == 0 {
		return alphabet[0:1]
	}

	negative := cmpZero < 0
	if negative {
		src.Abs(src)
	}

	var dst []byte
	radix := big.NewInt(int64(len(alphabet)))

	for src.Cmp(zero) > 0 {
		mod := new(big.Int)
		src.DivMod(src, radix, mod)

		dst = append(dst, alphabet[mod.Int64()])
	}

	for i, j := 0, len(dst)-1; i < j; i, j = i+1, j-1 {
		dst[i], dst[j] = dst[j], dst[i]
	}

	s := string(dst)
	if negative {
		s = "-" + s
	}

	return s
}
