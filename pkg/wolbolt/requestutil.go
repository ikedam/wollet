package wolbolt

import (
	"net"
	"net/http"
)

func getPath(r *http.Request) string {
	path := r.URL.Path
	if r.URL.RawPath != "" {
		path = r.URL.RawPath
	}
	if r.URL.RawQuery != "" {
		path += "?" + r.URL.RawQuery
	}
	return path
}

func getRemoteAddr(r *http.Request) string {
	remoteAddr := r.RemoteAddr
	if ip, _, err := net.SplitHostPort(remoteAddr); err == nil {
		remoteAddr = ip
	}
	return remoteAddr
}
