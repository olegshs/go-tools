package helpers

import (
	"net"
	"net/http"
	"strings"
)

func RequestIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); len(forwardedFor) > 0 {
		a := strings.Split(forwardedFor, ",")
		ip := Trim(a[0])
		return ip
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func IsAjax(r *http.Request) bool {
	if strings.ToLower(r.Header.Get("X-Requested-With")) == "xmlhttprequest" {
		return true
	}
	return false
}
