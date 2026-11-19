package typeconv

import (
	"bytes"
	"math"
	"testing"
	"time"
)

func TestTo(t *testing.T) {
	v, ok := To("string", 12345).(string)
	if !ok || v != "12345" {
		t.Error(`To("string", 12345) != "12345"`)
	}
}

func TestString(t *testing.T) {
	if String("abc") != "abc" {
		t.Error(`String("abc") != "abc"`)
	}

	if String([]byte{97, 98, 99}) != "abc" {
		t.Error(`String([]byte{97, 98, 99}) != "abc"`)
	}

	if String(-123) != "-123" {
		t.Error(`String(123) != "-123"`)
	}

	if String(math.Pi) != "3.141592653589793" {
		t.Error(`String(math.Pi) != "3.141592653589793"`)
	}

	if String(false) != "false" {
		t.Error(`String(false) != "false"`)
	}

	if String(true) != "true" {
		t.Error(`String(true) != "true"`)
	}
}

func TestBool(t *testing.T) {
	if Bool(true) != true {
		t.Error(`Bool(true) != true`)
	}

	if Bool("") != false {
		t.Error(`Bool("") != false`)
	}

	if Bool("0") != true {
		t.Error(`Bool("0") != true`)
	}

	if Bool(0) != false {
		t.Error(`Bool(0) != false`)
	}

	if Bool(-1) != true {
		t.Error(`Bool(-1) != true`)
	}
}

func TestInt(t *testing.T) {
	if Int(false) != 0 {
		t.Error(`Int(false) != 0`)
	}

	if Int(true) != 1 {
		t.Error(`Int(true) != 1`)
	}
}

func TestInt32(t *testing.T) {
	if Int32("-0xFFFFFFFF") != 1 {
		t.Error(`Int32("-0xFFFFFFFF") != 1`)
	}

	if Uint32("-1") != 0xFFFFFFFF {
		t.Error(`Uint32("-1") != 0xFFFFFFFF`)
	}
}

func TestInt64(t *testing.T) {
	if Int64("-0xFFFFFFFFFFFFFFF") != -0xFFFFFFFFFFFFFFF {
		t.Error(`Int64("-0xFFFFFFFFFFFFFFF") != -0xFFFFFFFFFFFFFFF`)
	}

	if Uint64("-1") != 0xFFFFFFFFFFFFFFFF {
		t.Error(`Uint64("-1") != 0xFFFFFFFFFFFFFFFF`)
	}
}

func TestFloat64(t *testing.T) {
	if Float64(false) != 0 {
		t.Error(`Float64(false) != 0`)
	}

	if Float64(true) != 1 {
		t.Error(`Float64(true) != 1`)
	}

	if Float64("3.141592653589793") != math.Pi {
		t.Error(`Float64("3.141592653589793") != math.Pi`)
	}
}

func TestBytes(t *testing.T) {
	if !bytes.Equal(Bytes("abc"), []byte{97, 98, 99}) {
		t.Error(`Bytes("abc") != []byte{97, 98, 99}`)
	}

	if !bytes.Equal(Bytes(int32(1)), []byte{1, 0, 0, 0}) {
		t.Error(`Bytes(int32(1)) != []byte{1, 0, 0, 0}`)
	}

	if !bytes.Equal(BytesBigEndian(int32(1)), []byte{0, 0, 0, 1}) {
		t.Error(`BytesBigEndian(int32(1)) != []byte{0, 0, 0, 1}`)
	}
}

func TestDuration(t *testing.T) {
	m := map[interface{}]time.Duration{
		"30s":      30 * time.Second,
		"1m30s":    90 * time.Second,
		"30m30s":   30*time.Minute + 30*time.Second,
		"1h30m30s": 1*time.Hour + 30*time.Minute + 30*time.Second,
		600:        600 * time.Second,
		12.345678:  12345678 * time.Microsecond,
	}

	for v, exp := range m {
		d := Duration(v)
		if d != exp {
			t.Errorf(`Duration(%v) == %v, expected: %v`, v, d, exp)
		}
	}
}

func TestTime(t *testing.T) {
	exp := time.Date(2006, 1, 2, 15, 4, 5, 999999999, time.UTC)

	m := map[interface{}]time.Duration{
		"2006-01-02T15:04:05.999Z": time.Millisecond,
		"2006-01-02 15:04:05.999":  time.Millisecond,
		"2006-01-02T15:04:05Z":     time.Second,
		"2006-01-02 15:04:05":      time.Second,
		1136214245:                 time.Second,
		1136214245.999999:          time.Microsecond,
	}

	for v, truncate := range m {
		ts := Time(v)
		e := exp.Truncate(truncate)
		if !ts.Equal(e) {
			t.Errorf(`Time(%v) == %v, expected: %v`, v, ts, e)
		}
	}
}
