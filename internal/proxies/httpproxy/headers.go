package httpproxy

import (
	"net/http"
	"strings"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// defined in: http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailer",
	"Transfer-Encoding",
	"Upgrade",
}

func cleanHeaders(hdr http.Header) http.Header {
	if c := hdr.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				hdr.Del(f)
			}
		}
	}

	for _, h := range hopHeaders {
		hdr.Del(h)
	}

	return hdr
}

func cloneCleanHeader(h http.Header) http.Header {
	x := cloneHeader(h)
	return cleanHeaders(x)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
