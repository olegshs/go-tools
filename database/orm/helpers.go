package orm

import (
	"reflect"
)

func fieldByIndex(v reflect.Value, index ...[]int) reflect.Value {
	for _, i := range index {
		v = v.FieldByIndex(i)
	}
	return v
}
