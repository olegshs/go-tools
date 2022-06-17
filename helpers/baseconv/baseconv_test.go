package baseconv

import (
	"strconv"
	"testing"
)

func TestFormatInt(t *testing.T) {
	i := int64(9223372036854775807)

	expected := strconv.FormatInt(i, 36)

	got := FormatInt(i, AlphabetBase36)
	if got != expected {
		t.Errorf("%d == %s₍₃₆₎ != %s₍₃₆₎", i, expected, got)
	}
}

func TestParse(t *testing.T) {
	s := "1y2p0ij32e8e7"

	expected := int64(9223372036854775807)

	b, err := Parse(s, AlphabetBase36)
	if err != nil {
		t.Error(err)
	}

	got := b.Int64()
	if got != expected {
		t.Errorf("%s₍₃₆₎ == %d != %d", s, expected, got)
	}
}
