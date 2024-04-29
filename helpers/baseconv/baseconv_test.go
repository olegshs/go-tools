package baseconv

import (
	"strconv"
	"testing"
)

func TestFormatInt(t *testing.T) {
	a := []int64{
		-9223372036854775808,
		-1,
		0,
		1,
		9223372036854775807,
	}

	for _, v := range a {
		expected := strconv.FormatInt(v, 36)

		got := FormatInt(v, AlphabetBase36)
		if got != expected {
			t.Errorf("%d == %s₍₃₆₎ != %s₍₃₆₎", v, expected, got)
		}
	}
}

func TestParse(t *testing.T) {
	a := []string{
		"-1y2p0ij32e8e8",
		"-1",
		"0",
		"1",
		"1y2p0ij32e8e7",
	}

	for _, v := range a {
		expected, err := strconv.ParseInt(v, 36, 64)
		if err != nil {
			t.Error(err)
		}

		b, err := Parse(v, AlphabetBase36)
		if err != nil {
			t.Error(err)
		}

		got := b.Int64()
		if got != expected {
			t.Errorf("%s₍₃₆₎ == %d != %d", v, expected, got)
		}
	}
}
