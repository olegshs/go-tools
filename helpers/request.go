package helpers

import (
	"net"
	"net/http"
	"strings"
)

func RequestIP(r *http.Request) string {
	xForwardedFor := []string{
		"X-Forwarded-For",
		"x-forwarded-for",
		"X-FORWARDED-FOR",
	}
	for _, key := range xForwardedFor {
		if value := r.Header.Get(key); len(value) > 0 {
			a := strings.Split(value, ",")
			ip := Trim(a[0])
			return ip
		}
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func IsAjax(r *http.Request) bool {
	xForwardedFor := []string{
		"X-Requested-With",
		"x-requested-with",
		"X-REQUESTED-WITH",
	}
	for _, key := range xForwardedFor {
		if strings.ToLower(r.Header.Get(key)) == "xmlhttprequest" {
			return true
		}
	}
	return false
}
