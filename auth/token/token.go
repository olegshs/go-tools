// Пакет token предоставляет функции для кодирования, декодирования и проверки аутентификационных токенов.
package token

import (
	"bytes"
	"crypto"
	"encoding/binary"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	bytesKind = reflect.TypeOf([]byte{}).Kind()
	timeKind  = reflect.TypeOf(time.Time{}).Kind()

	kindLengths = map[reflect.Kind]int{
		reflect.Bool:   1,
		reflect.Int8:   1,
		reflect.Uint8:  1,
		reflect.Int16:  2,
		reflect.Uint16: 2,
		reflect.Int32:  4,
		reflect.Uint32: 4,
		reflect.Int64:  8,
		reflect.Uint64: 8,
		reflect.String: 2,
		bytesKind:      2,
		timeKind:       8,
	}
)

func Encode(payload interface{}, hash crypto.Hash, secrets ...interface{}) ([]byte, error) {
	value := reflect.ValueOf(payload)
	for value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return nil, errors.New("payload is not a struct")
	}

	buf := new(bytes.Buffer)

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		err := encodeField(buf, field)
		if err != nil {
			return nil, err
		}
	}

	signature, err := generateSignature(buf, hash, secrets...)
	if err != nil {
		return nil, err
	}

	buf.Write(signature)

	return buf.Bytes(), nil
}

func Decode(token []byte, payload interface{}) error {
	value := reflect.ValueOf(payload)
	if value.Kind() != reflect.Ptr {
		return errors.New("payload is not a pointer")
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return errors.New("payload is not a struct")
	}

	buf := bytes.NewBuffer(token)

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		err := decodeField(buf, field)
		if err != nil {
			return err
		}
	}

	return nil
}

func Validate(token []byte, hash crypto.Hash, secrets ...interface{}) error {
	payloadSize := len(token) - hash.Size()
	if payloadSize < 0 {
		return errors.New("unexpected end")
	}

	payload := token[:payloadSize]
	signature := token[payloadSize:]

	buf := new(bytes.Buffer)
	buf.Write(payload)

	validSignature, err := generateSignature(buf, hash, secrets...)
	if err != nil {
		return err
	}

	if bytes.Compare(signature, validSignature) != 0 {
		return errors.New("invalid signature")
	}
	return nil
}

func encodeField(buf *bytes.Buffer, field reflect.Value) error {
	kind := field.Kind()

	n, ok := kindLengths[kind]
	if !ok {
		return errorUnsupportedType(kind)
	}

	b := make([]byte, n)

	switch kind {
	case reflect.Bool:
		if field.Bool() {
			b[0] = 1
		} else {
			b[0] = 0
		}

	case reflect.Int8:
		b[0] = byte(field.Int())

	case reflect.Uint8:
		b[0] = byte(field.Uint())

	case reflect.Int16:
		i := uint16(field.Int())
		binary.LittleEndian.PutUint16(b, i)

	case reflect.Uint16:
		i := uint16(field.Uint())
		binary.LittleEndian.PutUint16(b, i)

	case reflect.Int32:
		i := uint32(field.Int())
		binary.LittleEndian.PutUint32(b, i)

	case reflect.Uint32:
		i := uint32(field.Uint())
		binary.LittleEndian.PutUint32(b, i)

	case reflect.Int64:
		i := uint64(field.Int())
		binary.LittleEndian.PutUint64(b, i)

	case reflect.Uint64:
		i := field.Uint()
		binary.LittleEndian.PutUint64(b, i)

	case timeKind:
		t := field.Interface().(time.Time)
		i := uint64(t.UnixNano())
		binary.LittleEndian.PutUint64(b, i)

	case reflect.String, bytesKind:
		if kind == reflect.String {
			b = []byte(field.String())
		} else {
			b = field.Bytes()
		}

		n := len(b)
		if n >= 0x10000 {
			return errors.New("string is too long")
		}

		n2 := make([]byte, 2)
		binary.LittleEndian.PutUint16(n2, uint16(n))

		b = append(n2, b...)

	default:
		return errorUnsupportedType(kind)
	}

	buf.Write(b)

	return nil
}

func decodeField(buf *bytes.Buffer, field reflect.Value) error {
	kind := field.Kind()

	n, ok := kindLengths[kind]
	if !ok {
		return errorUnsupportedType(kind)
	}
	if n > buf.Len() {
		return errors.New("unexpected end")
	}

	b := make([]byte, n)
	n, err := buf.Read(b)
	if err != nil {
		return err
	}

	switch kind {
	case reflect.Bool:
		field.SetBool(b[0] != 0)

	case reflect.Int8, reflect.Uint8:
		if kind == reflect.Int8 {
			field.SetInt(int64(b[0]))
		} else {
			field.SetUint(uint64(b[0]))
		}

	case reflect.Int16, reflect.Uint16:
		i := binary.LittleEndian.Uint16(b)
		if kind == reflect.Int16 {
			field.SetInt(int64(i))
		} else {
			field.SetUint(uint64(i))
		}

	case reflect.Int32, reflect.Uint32:
		i := binary.LittleEndian.Uint32(b)
		if kind == reflect.Int32 {
			field.SetInt(int64(i))
		} else {
			field.SetUint(uint64(i))
		}

	case reflect.Int64, reflect.Uint64:
		i := binary.LittleEndian.Uint64(b)
		if kind == reflect.Int64 {
			field.SetInt(int64(i))
		} else {
			field.SetUint(i)
		}

	case timeKind:
		i := int64(binary.LittleEndian.Uint64(b))
		t := time.Unix(i/1000000000, i%1000000000)
		field.Set(reflect.ValueOf(t.UTC()))

	case reflect.String, bytesKind:
		n = int(binary.LittleEndian.Uint16(b))
		if n > buf.Len() {
			return errors.New("unexpected end")
		}

		b = make([]byte, n)
		n, err = buf.Read(b)
		if err != nil {
			return err
		}

		if kind == bytesKind {
			field.SetBytes(b)
		} else {
			field.SetString(string(b))
		}

	default:
		return errorUnsupportedType(kind)
	}

	return nil
}

func generateSignature(buf *bytes.Buffer, hash crypto.Hash, secrets ...interface{}) ([]byte, error) {
	if !hash.Available() {
		return nil, errors.New("hash is not available")
	}

	b, err := secretsToBytes(secrets...)
	if err != nil {
		return nil, err
	}

	h := hash.New()
	h.Write(buf.Bytes())
	h.Write(b)

	signature := h.Sum(nil)
	return signature, nil
}

func secretsToBytes(secrets ...interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)

	for _, v := range secrets {
		var b []byte

		switch t := v.(type) {
		case bool:
			b = make([]byte, 1)
			if t {
				b[0] = 1
			} else {
				b[0] = 0
			}

		case int8:
			b = make([]byte, 1)
			b[0] = byte(t)

		case uint8:
			b = make([]byte, 1)
			b[0] = byte(t)

		case int16:
			b = make([]byte, 2)
			binary.LittleEndian.PutUint16(b, uint16(t))

		case uint16:
			b = make([]byte, 2)
			binary.LittleEndian.PutUint16(b, t)

		case int32:
			b = make([]byte, 4)
			binary.LittleEndian.PutUint32(b, uint32(t))

		case uint32:
			b = make([]byte, 4)
			binary.LittleEndian.PutUint32(b, t)

		case int64:
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(t))

		case uint64:
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, t)

		case int:
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(t))

		case uint:
			b = make([]byte, 8)
			binary.LittleEndian.PutUint64(b, uint64(t))

		case string:
			b = []byte(t)

		case []byte:
			b = t

		default:
			return nil, errorUnsupportedType(reflect.TypeOf(v).Kind())
		}

		buf.Write(b)
	}

	return buf.Bytes(), nil
}

func errorUnsupportedType(kind reflect.Kind) error {
	return errors.New(fmt.Sprintf("unsupported type: %s", kind))
}
