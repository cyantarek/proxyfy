package httpproxy

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

// Format a time duration
func format(t time.Duration) string {
	u0 := t.Nanoseconds() / 1000
	ma, mf := u0/1000, u0%1000

	if ma == 0 {
		return fmt.Sprintf("%3.3d us", mf)
	}

	return fmt.Sprintf("%d.%3.3d ms", ma, mf)
}

func extractHost(u *url.URL) string {
	h := u.Host

	i := strings.LastIndex(h, ":")
	if i < 0 {
		h += ":80"
	}
	return h
}
