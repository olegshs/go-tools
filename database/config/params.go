package config

import (
	"net/url"

	"github.com/olegshs/go-tools/helpers/typeconv"
)

// Params holds extra connection parameters that drivers append to the DSN.
type Params map[string]interface{}

// String returns the parameters as a URL query string, including the
// leading "?", or "" if there are none.
func (p Params) String() string {
	if len(p) == 0 {
		return ""
	}

	a := url.Values{}
	for k, v := range p {
		a.Add(k, typeconv.String(v))
	}

	s := "?" + a.Encode()
	return s
}
