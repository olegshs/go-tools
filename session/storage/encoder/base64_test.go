package encoder

import (
	"bytes"
	"testing"
)

func TestBase64_RoundTrip(t *testing.T) {
	encoder := new(Base64)

	for _, s := range []string{"", "a", "ab", "abc", "abcd"} {
		encoded, err := encoder.Encode([]byte(s))
		if err != nil {
			t.Fatal(err)
		}

		decoded, err := encoder.Decode(encoded)
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(decoded, []byte(s)) {
			t.Errorf("Decode(Encode(%q)) = %q", s, decoded)
		}
	}
}
