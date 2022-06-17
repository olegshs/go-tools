package token

import (
	"crypto"
	_ "crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

type Payload struct {
	Bool   bool
	Int8   int8
	Uint8  uint8
	Int16  int16
	Uint16 uint16
	Int32  int32
	Uint32 uint32
	Int64  int64
	Uint64 uint64
	Time   time.Time
	String string
	Bytes  []byte
}

func TestToken(test *testing.T) {
	secrets := []interface{}{
		true,
		int8(-128),
		uint8(255),
		int16(-32768),
		uint16(65535),
		int32(-2147483648),
		uint32(4294967295),
		int64(-9223372036854775808),
		uint64(18446744073709551615),
		"test",
		[]byte{1, 2, 3, 4},
	}

	s, err := secretsToBytes(secrets...)
	if err != nil {
		test.Error("secretsToBytes:", err)
	}

	validBase64 := "AYD_AID__wAAAID_____AAAAAAAAAID__________3Rlc3QBAgME"
	if base64.RawURLEncoding.EncodeToString(s) != validBase64 {
		test.Error("secretsToBytes:", "invalid result")
	}

	t1 := Payload{
		true,
		127,
		255,
		32767,
		65535,
		-2147483648,
		4294967295,
		-9223372036854775808,
		18446744073709551615,
		time.Date(2018, 06, 30, 1, 2, 3, 4, time.UTC),
		"This is a test.",
		[]byte{1, 2, 3, 4, 5},
	}

	b, err := Encode(t1, crypto.SHA256, secrets...)
	if err != nil {
		test.Error("Encode:", err)
	}

	validBase64 = "AX___3___wAAAID_____AAAAAAAAAID__________wSu511oyjwVDwBUaG" +
		"lzIGlzIGEgdGVzdC4FAAECAwQFXCw9fL5vwB7qKf0kiOHy3YrQggrH-ezhnKSjFvu4SHE"
	if base64.RawURLEncoding.EncodeToString(b) != validBase64 {
		test.Error("Encode:", "invalid result")
	}

	t2 := Payload{}
	err = Decode(b, &t2)
	if err != nil {
		test.Error("Decode:", err)
	}

	if fmt.Sprint(t1) != fmt.Sprint(t2) {
		test.Error(t1, "!=", t2)
	}

	err = Validate(b, crypto.SHA256, secrets...)
	if err != nil {
		test.Error("Validate:", err)
	}

	err = Validate(b, crypto.SHA256, secrets[:len(secrets)-1]...)
	if err == nil {
		test.Error("Validate:", "no error for invalid secrets")
	}
}
