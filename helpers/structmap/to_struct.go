// Package structmap copies loosely typed data, such as decoded JSON or YAML
// maps, into typed structs and slices.
package structmap

import (
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/olegshs/go-tools/helpers"
	"github.com/olegshs/go-tools/helpers/typeconv"
)

// ToStruct copies values from src into the struct that dst points to.
//
// For each field, the map key is looked up in this order: the name in the
// field's json tag, the Go field name, and the field name in snake_case.
// Fields tagged `json:"-"` are skipped. Embedded structs are filled from the
// same map. Nested maps fill nested structs, and values are converted to the
// field type with typeconv. Fields without a matching key keep their value,
// so dst can be pre-filled with defaults. Values that cannot be converted
// are silently ignored.
func ToStruct(src map[string]interface{}, dst interface{}) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		return
	}

	toStruct(src, rv)
}

// ToSlice converts each element of src to the element type of the slice that
// dst points to, using the same rules as ToStruct, and stores the result.
func ToSlice(src []interface{}, dst interface{}) {
	rv := reflect.ValueOf(dst)
	if rv.Kind() != reflect.Ptr {
		return
	}

	toSlice(src, rv)
}

func toStruct(src map[string]interface{}, dst reflect.Value) {
	dst = elem(dst)
	if dst.Kind() != reflect.Struct {
		return
	}

	dstType := dst.Type()

	n := dst.NumField()
	for i := 0; i < n; i++ {
		f := dst.Field(i)
		t := dstType.Field(i)

		if t.Anonymous {
			toStruct(src, f)
			continue
		}

		v := mapValue(src, t)
		if v == nil {
			continue
		}

		to(v, elem(f))
	}
}

func toSlice(src interface{}, dst reflect.Value) {
	dst = elem(dst)
	if dst.Kind() != reflect.Slice {
		return
	}

	dstType := dst.Type()
	itemType := dstType.Elem()

	rv := reflect.ValueOf(src)
	// A single value of the element type becomes a one-element slice.
	if (rv.Kind() != reflect.Slice) && (rv.Type() == itemType) {
		a := reflect.MakeSlice(dstType, 1, 1)
		toValue(src, a.Index(0))
		dst.Set(a)
		return
	}

	n := rv.Len()
	a := reflect.MakeSlice(dstType, n, n)
	for i := 0; i < n; i++ {
		item := rv.Index(i)

		newItem := reflect.New(itemType).Elem()
		to(item.Interface(), elem(newItem))

		a.Index(i).Set(newItem)
	}

	dst.Set(a)
}

func to(src interface{}, dst reflect.Value) {
	switch dst.Kind() {
	case reflect.Struct:
		if m, ok := src.(map[string]interface{}); ok {
			toStruct(m, dst)
		} else {
			toValue(src, dst)
		}

	case reflect.Slice:
		toSlice(src, dst)

	default:
		toValue(src, dst)
	}
}

// toValue stores src in dst, converting it with typeconv or reflection if the
// types differ. Values that cannot be converted are skipped.
func toValue(src interface{}, dst reflect.Value) {
	dstType := dst.Type()

	rv := reflect.ValueOf(src)
	t := rv.Type()

	if t != dstType {
		src = typeconv.To(dstType.String(), src)
		if src != nil {
			rv = reflect.ValueOf(src)
		} else if t.ConvertibleTo(dstType) {
			rv = rv.Convert(dstType)
		} else {
			return
		}
	}

	dst.Set(rv)
}

// elem dereferences v, allocating nil pointers along the way.
func elem(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

// mapValue returns the value in m for field t (see ToStruct for the lookup
// order), or nil if there is none.
func mapValue(m map[string]interface{}, t reflect.StructField) interface{} {
	tag, ok := t.Tag.Lookup("json")
	if ok {
		a := strings.Split(tag, ",")
		k := helpers.Trim(a[0])

		switch k {
		case "":
			k = t.Name
		case "-":
			return nil
		}

		if v, ok := m[k]; ok {
			return v
		}
	}

	if v, ok := m[t.Name]; ok {
		return v
	}

	k := strcase.ToSnake(t.Name)
	if v, ok := m[k]; ok {
		return v
	}

	return nil
}
