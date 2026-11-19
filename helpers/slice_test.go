package helpers

import (
	"testing"
)

func TestSlice_IndexOf(t *testing.T) {
	a := Slice[string]{"a", "b", "c"}

	if a.IndexOf("c") != 2 {
		t.Errorf(`%v: index of "c" != 2`, a)
	}

	if a.IndexOf("x") != -1 {
		t.Errorf(`%v: index of "x" != -1`, a)
	}
}

func TestSlice_Map(t *testing.T) {
	a := Slice[string]{"a", "b", "c"}

	b := a.Map(func(s string) string {
		return s + "!"
	})

	for i, v := range b {
		s := a[i] + "!"
		if v != s {
			t.Errorf(`%v: element %d != "%s"`, b, i, s)
		}
	}
}
