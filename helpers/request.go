package helpers

import (
	"net"
	"net/http"
	"strings"
)

// RequestIP returns the client IP address of the request. It uses the first
// address in the X-Forwarded-For header if present, and otherwise the host
// part of r.RemoteAddr.
//
// X-Forwarded-For is set by the client and can be forged, so only rely on it
// behind a trusted reverse proxy that overwrites the header.
func RequestIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); len(forwardedFor) > 0 {
		a := strings.Split(forwardedFor, ",")
		ip := Trim(a[0])
		return ip
	}

	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// IsAjax reports whether the request was sent by XMLHttpRequest, based on the
// X-Requested-With header that many JavaScript libraries set.
func IsAjax(r *http.Request) bool {
	if strings.ToLower(r.Header.Get("X-Requested-With")) == "xmlhttprequest" {
		return true
	}
	return false
}
