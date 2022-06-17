package config

import (
	"net/url"

	"github.com/olegshs/go-tools/helpers/typeconv"
)

type Params map[string]interface{}

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
