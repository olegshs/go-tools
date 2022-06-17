package router

import (
	"context"
	"net/http"
)

type Params []Param

type Param struct {
	Key   string
	Value string
}

type paramsKeyType struct{}

var paramsKey = paramsKeyType{}

// ParamsFromRequest получает структуру с именованными параметрами из HTTP запроса.
func ParamsFromRequest(r *http.Request) Params {
	params, _ := r.Context().Value(paramsKey).(Params)
	return params
}

// ByName возвращает значение параметра по его имени.
func (params Params) ByName(name string) string {
	for _, p := range params {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

// Values возвращает массив строк со значениями всех параметров.
func (params Params) Values() []string {
	a := make([]string, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// InterfaceValues возвращает массив пустых интерфейсов со значениями всех параметров.
func (params Params) InterfaceValues() []interface{} {
	a := make([]interface{}, len(params))
	for i, p := range params {
		a[i] = p.Value
	}
	return a
}

// Map возвращает карту всех параметров. Значения - строки.
func (params Params) Map() map[string]string {
	m := make(map[string]string, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

// InterfaceMap возвращает карту всех параметров. Значения - пустые интерфейсы.
func (params Params) InterfaceMap() map[string]interface{} {
	m := make(map[string]interface{}, len(params))
	for _, p := range params {
		m[p.Key] = p.Value
	}
	return m
}

func (params Params) toRequest(r *http.Request) {
	ctx := r.Context()
	ctx = context.WithValue(ctx, paramsKey, params)
	*r = *r.WithContext(ctx)
}
