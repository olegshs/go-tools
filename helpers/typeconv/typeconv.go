// Package typeconv converts loosely typed values (interface{}) to concrete
// Go types without returning errors. Values that cannot be converted produce
// the zero value of the target type.
//
// It accepts all built-in numeric types, bool, string, []byte, json.Number,
// time.Time and *time.Time. Strings are parsed leniently, so "42px" becomes
// 42. It is used to decode configuration and database values.
package typeconv

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

const (
	is32 = strconv.IntSize == 32

	// Time layouts used when parsing and formatting time values.
	timeFormatSql     = "2006-01-02 15:04:05.999"
	timeFormatRFC3339 = "2006-01-02T15:04:05.999Z07:00"
)

var (
	// Patterns that extract the first number from a string when strict
	// parsing fails.
	stringToIntRegexp   = regexp.MustCompile(`(0x)?\d+`)
	stringToFloatRegexp = regexp.MustCompile(`\d+(\.\d+)?`)

	timeFormatSqlRegexp = regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}(\.\d+)?`)
)

// CanTo reports whether To supports the type with the given name, as
// returned by reflect.Type.String (e.g. "int64", "time.Duration").
func CanTo(t string) bool {
	switch t {
	case "string", "bool", "int", "uint", "int8", "uint8", "int16", "uint16",
		"int32", "uint32", "int64", "uint64", "float32", "float64", "[]uint8",
		"time.Duration", "time.Time", "*time.Time":
		return true
	default:
		return false
	}
}

// To converts value to the type with the given name. It returns nil if the
// type is not supported (see CanTo).
func To(t string, value interface{}) interface{} {
	switch t {
	case "string":
		return String(value)
	case "bool":
		return Bool(value)
	case "int":
		return Int(value)
	case "uint":
		return Uint(value)
	case "int8":
		return Int8(value)
	case "uint8":
		return Uint8(value)
	case "int16":
		return Int16(value)
	case "uint16":
		return Uint16(value)
	case "int32":
		return Int32(value)
	case "uint32":
		return Uint32(value)
	case "int64":
		return Int64(value)
	case "uint64":
		return Uint64(value)
	case "float32":
		return Float32(value)
	case "float64":
		return Float64(value)
	case "[]uint8":
		return Bytes(value)
	case "time.Time":
		return Time(value)
	case "*time.Time":
		return TimePtr(value)
	case "time.Duration":
		return Duration(value)
	default:
		return nil
	}
}

// String converts value to a string. Times use RFC 3339 format, and nil
// becomes the empty string.
func String(value interface{}) string {
	switch t := value.(type) {
	case bool:
		if t {
			return "true"
		} else {
			return "false"
		}
	case int, int8, int16, int32, int64:
		return strconv.FormatInt(Int64(value), 10)
	case uint, uint8, uint16, uint32, uint64:
		return strconv.FormatUint(Uint64(value), 10)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case string:
		return t
	case []byte:
		return string(t)
	case time.Time:
		return t.Format(timeFormatRFC3339)
	case *time.Time:
		if t == nil {
			return ""
		}
		return t.Format(timeFormatRFC3339)
	default:
		if t == nil {
			return ""
		}
		return fmt.Sprint(t)
	}
}

// Bool converts value to a bool. Any non-empty string is true, including
// "false" and "0". Other values are true when their integer value is not
// zero.
func Bool(value interface{}) bool {
	switch t := value.(type) {
	case bool:
		return t
	case string:
		return len(t) > 0
	default:
		i := Int64(value)
		return i != 0
	}
}

// Int converts value to an int. See Int64 for the conversion rules.
func Int(value interface{}) int {
	if is32 {
		return int(Int32(value))
	} else {
		return int(Int64(value))
	}
}

// Uint converts value to a uint. Negative values wrap around.
func Uint(value interface{}) uint {
	return uint(Int(value))
}

// Int8 converts value to an int8, truncating out-of-range values.
func Int8(value interface{}) int8 {
	return int8(Int32(value))
}

// Uint8 converts value to a uint8, truncating out-of-range values.
func Uint8(value interface{}) uint8 {
	return uint8(Int32(value))
}

// Int16 converts value to an int16, truncating out-of-range values.
func Int16(value interface{}) int16 {
	return int16(Int32(value))
}

// Uint16 converts value to a uint16, truncating out-of-range values.
func Uint16(value interface{}) uint16 {
	return uint16(Int32(value))
}

// Int32 converts value to an int32, truncating out-of-range values.
// See Int64 for the conversion rules.
func Int32(value interface{}) int32 {
	switch t := value.(type) {
	case bool:
		if t {
			return 1
		} else {
			return 0
		}
	case int:
		return int32(t)
	case uint:
		return int32(t)
	case int8:
		return int32(t)
	case uint8:
		return int32(t)
	case int16:
		return int32(t)
	case uint16:
		return int32(t)
	case int32:
		return t
	case uint32:
		return int32(t)
	case int64:
		return int32(t)
	case uint64:
		return int32(t)
	case float32:
		return int32(t)
	case float64:
		return int32(t)
	case string:
		return int32(StringToInt(t))
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0
		}
		return int32(i)
	default:
		return 0
	}
}

// Uint32 converts value to a uint32, truncating out-of-range values.
func Uint32(value interface{}) uint32 {
	return uint32(Int32(value))
}

// Int64 converts value to an int64. Booleans become 0 or 1, floats are
// truncated toward zero, and strings are parsed with StringToInt.
// Unsupported types return 0.
func Int64(value interface{}) int64 {
	switch t := value.(type) {
	case bool:
		if t {
			return 1
		} else {
			return 0
		}
	case int:
		return int64(t)
	case uint:
		return int64(t)
	case int8:
		return int64(t)
	case uint8:
		return int64(t)
	case int16:
		return int64(t)
	case uint16:
		return int64(t)
	case int32:
		return int64(t)
	case uint32:
		return int64(t)
	case int64:
		return t
	case uint64:
		return int64(t)
	case float32:
		return int64(t)
	case float64:
		return int64(t)
	case string:
		return StringToInt(t)
	case json.Number:
		i, err := t.Int64()
		if err != nil {
			return 0
		}
		return i
	default:
		return 0
	}
}

// Uint64 converts value to a uint64. Negative values wrap around.
func Uint64(value interface{}) uint64 {
	return uint64(Int64(value))
}

// Float32 converts value to a float32. See Float64 for the conversion rules.
func Float32(value interface{}) float32 {
	return float32(Float64(value))
}

// Float64 converts value to a float64. Booleans become 0 or 1, and strings
// are parsed with StringToFloat. Unsupported types return 0.
func Float64(value interface{}) float64 {
	switch t := value.(type) {
	case bool:
		if t {
			return 1
		} else {
			return 0
		}
	case int:
		return float64(t)
	case uint:
		return float64(t)
	case int8:
		return float64(t)
	case uint8:
		return float64(t)
	case int16:
		return float64(t)
	case uint16:
		return float64(t)
	case int32:
		return float64(t)
	case uint32:
		return float64(t)
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case float32:
		return float64(t)
	case float64:
		return t
	case string:
		return StringToFloat(t)
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return f
	default:
		return 0
	}
}

// Bytes converts value to bytes. Strings and byte slices are returned as is.
// Fixed-size numbers are encoded in little-endian order, and anything else
// is converted with String first.
func Bytes(value interface{}) []byte {
	return toBytes(value, binary.LittleEndian)
}

// BytesBigEndian is like Bytes but encodes numbers in big-endian order.
func BytesBigEndian(value interface{}) []byte {
	return toBytes(value, binary.BigEndian)
}

// toBytes implements Bytes and BytesBigEndian. int and uint have no fixed
// size, so they are first converted to the platform's 32- or 64-bit type.
func toBytes(value interface{}, order binary.ByteOrder) []byte {
	switch t := value.(type) {
	case []byte:
		return t
	case string:
		return []byte(t)
	case int:
		if is32 {
			value = int32(t)
		} else {
			value = int64(t)
		}
	case uint:
		if is32 {
			value = uint32(t)
		} else {
			value = uint64(t)
		}
	}

	if binary.Size(value) == -1 {
		return []byte(String(value))
	}

	buf := new(bytes.Buffer)

	err := binary.Write(buf, order, value)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}

// Duration converts value to a time.Duration. Numbers are seconds (floats
// keep microsecond precision), and strings are parsed with
// time.ParseDuration (e.g. "1m30s"). Invalid strings return 0.
func Duration(value interface{}) time.Duration {
	switch t := value.(type) {
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64:
		i := Int64(t)
		return time.Duration(i) * time.Second
	case float32, float64, json.Number:
		i, f := math.Modf(Float64(t))
		f = math.Round(f*1000000) / 1000000
		return time.Duration(i)*time.Second + time.Duration(f*1000000000)
	default:
		s := String(t)
		d, _ := time.ParseDuration(s)
		return d
	}
}

// Time converts value to a UTC time.Time. Numbers are Unix timestamps in
// seconds. Strings may be SQL datetimes ("2006-01-02 15:04:05") or RFC 3339.
// Invalid values return the zero time.
func Time(value interface{}) time.Time {
	ts, _ := toTime(value)
	ts = ts.UTC()
	return ts
}

// TimePtr is like Time but returns nil if value cannot be converted.
func TimePtr(value interface{}) *time.Time {
	ts, err := toTime(value)
	if err != nil {
		return nil
	}

	ts = ts.UTC()
	return &ts
}

// toTime implements Time and TimePtr. SQL datetimes without a zone are
// treated as UTC.
func toTime(value interface{}) (time.Time, error) {
	switch t := value.(type) {
	case int, uint, int32, uint32, int64, uint64:
		i := Int64(t)
		ts := time.Unix(i, 0)
		return ts, nil
	case float32, float64, json.Number:
		i, f := math.Modf(Float64(t))
		ts := time.Unix(int64(i), int64(f*1000000000)).Round(time.Microsecond)
		return ts, nil
	default:
		s := String(t)
		if timeFormatSqlRegexp.MatchString(s) {
			return time.Parse(timeFormatSql, s)
		} else {
			return time.Parse(time.RFC3339Nano, s)
		}
	}
}

// StringToInt parses s as an integer. Prefixes like 0x, 0o and 0b select the
// base, and a leading 0 means octal. If s is not a valid number, the first
// run of digits in it is used, so the sign is lost ("-5px" gives 5).
// It returns 0 if s contains no digits.
func StringToInt(s string) int64 {
	i, err := strconv.ParseInt(s, 0, 64)
	if err == nil {
		return i
	}

	a := stringToIntRegexp.FindString(s)
	if len(a) == 0 {
		return 0
	}

	i, err = strconv.ParseInt(a, 0, 64)
	if err != nil {
		return 0
	}

	return i
}

// StringToFloat parses s as a float. If s is not a valid number, the first
// decimal number in it is used, without the sign. It returns 0 if s contains
// no digits.
func StringToFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}

	a := stringToFloatRegexp.FindString(s)
	if len(a) == 0 {
		return 0
	}

	f, err = strconv.ParseFloat(a, 64)
	if err != nil {
		return 0
	}

	return f
}
