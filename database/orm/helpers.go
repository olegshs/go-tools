package orm

import (
	"reflect"
)

// fieldByIndex follows several field index paths in turn, for fields of
// embedded structs.
func fieldByIndex(v reflect.Value, index ...[]int) reflect.Value {
	for _, i := range index {
		v = v.FieldByIndex(i)
	}
	return v
}
