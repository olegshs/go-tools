package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func Dump(v interface{}) {
	Fdump(os.Stdout, v)
}

func Sdump(v interface{}) string {
	w := new(bytes.Buffer)
	Fdump(w, v)
	return w.String()
}

func Fdump(w io.Writer, v interface{}) {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
