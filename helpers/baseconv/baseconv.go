// Пакет baseconv предоставляет функции для преобразования данных
// при помощи различных алфавитов кодирования.
package baseconv

import (
	"errors"
	"fmt"
	"math/big"
)

const (
	AlphabetBase36 = "0123456789abcdefghijklmnopqrstuvwxyz"
	AlphabetBase58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	AlphabetBase62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

func Parse(src string, alphabet string) (*big.Int, error) {
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
			e := fmt.Sprintf("Invalid character at offset %d", i)
			return nil, errors.New(e)
		}

		dst.Mul(dst, radix)
		dst.Add(dst, big.NewInt(int64(b)))
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
	var dst []byte
	radix := big.NewInt(int64(len(alphabet)))
	zero := big.NewInt(0)

	for src.Cmp(zero) > 0 {
		mod := new(big.Int)
		src.DivMod(src, radix, mod)

		dst = append(dst, alphabet[mod.Int64()])
	}

	for i, j := 0, len(dst)-1; i < j; i, j = i+1, j-1 {
		dst[i], dst[j] = dst[j], dst[i]
	}

	return string(dst)
}
