package router

import (
	"context"
	"net/http"
)

// Params holds the path parameters of a matched route, in the order they
// appear in the pattern.
type Params []Param

// Param is a single path parameter.
type Param struct {
	Key   string
	Value string
}

// paramsKeyType is the context key type for the request's Params.
type paramsKeyType struct{}

var paramsKey = paramsKeyType{}

// ParamsFromRequest returns the path parameters of the route that handles r,
// or nil if the route has none.
func ParamsFromRequest(r *http.Request) Params {
	params, _ := r.Context().Value(paramsKey).(Params)
	return params
}

// ByName returns the value of the named parameter, or "" if there is none.
func (params Params) ByName(name string) string {
	for _, p := range params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Values returns the parameter values in order.
func (params Params) Values() []string {
	a := make([]string, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// InterfaceValues returns the parameter values in order as []interface{},
// for passing to variadic functions such as Router.Url.
func (params Params) InterfaceValues() []interface{} {
	a := make([]interface{}, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// Map returns the parameters as a name to value map.
func (params Params) Map() map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

// InterfaceMap returns the parameters as a name to value map with
// interface{} values, which is convenient for template data.
func (params Params) InterfaceMap() map[string]interface{} {
	m := make(map[string]interface{}, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

// toRequest stores params in the request context. It replaces *r in place
// so the caller's request carries the new context.
func (params Params) toRequest(r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, paramsKey, params)
	*r = *r.WithContext(ctx)
}
