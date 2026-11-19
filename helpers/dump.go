package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

// Dump prints v to stdout as indented JSON. It is meant for debugging.
func Dump(v interface{}) {
	Fdump(os.Stdout, v)
}

// Sdump returns v formatted as indented JSON.
func Sdump(v interface{}) string {
	w := new(bytes.Buffer)
	Fdump(w, v)
	return w.String()
}

// Fdump writes v to w as indented JSON. Encoding errors are ignored.
func Fdump(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
