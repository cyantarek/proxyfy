package httpproxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"proxyfy/pkg/copier"
	"proxyfy/pkg/privileges"
)

func (p *HTTPProxy) Accept() (net.Conn, error) {
	ln := p.lis
	for {
		nc, err := ln.Accept()
		select {
		case <-p.ctx.Done():
			return nil, &errShutdown

		default:
		}

		if err != nil {
			if ne, ok := err.(net.Error); ok {
				if ne.Timeout() || ne.Temporary() {
					continue
				}
			}
			return nil, err
		}

		if !p.limiter.Allow() {
			nc.Close()
			log.Printf("%s: globally rate limited\n", nc.RemoteAddr().String())
			continue
		}

		if !p.limiter.AllowHost(nc.RemoteAddr()) {
			nc.Close()
			log.Printf("%s: per-IP rate limited\n", nc.RemoteAddr().String())
			continue
		}

		if !privileges.ACLCheck(p.conf, nc) {
			log.Printf("%s: ACL failure\n", nc.RemoteAddr().String())
			nc.Close()
			continue
		}

		return nc, nil
	}
}

// handleConnect handles HTTP CONNECT
func (p *HTTPProxy) handleConnect(w http.ResponseWriter, r *http.Request) {
	h, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("can't do CONNECT: hijack failed\n")
		http.Error(w, "Can't support CONNECT", http.StatusNotImplemented)

		return
	}

	client, _, err := h.Hijack()
	if err != nil {
		log.Printf("can't do CONNECT: hijack failed: %s\n", err)
		http.Error(w, "Can't support CONNECT", http.StatusNotImplemented)
		client.Close()

		return
	}

	host := extractHost(r.URL)

	ctx := r.Context()

	dest, err := p.transport.DialContext(ctx, "tcp", host)
	if err != nil {
		log.Printf("can't connect to %s: %s\n", host, err)
		http.Error(w, fmt.Sprintf("can't connect to %s", host), http.StatusInternalServerError)
		client.Close()
		return
	}

	// since this is a hijacked connection, we need to write HTTP 200 OK at the beginning of the response
	client.Write(_200Ok)

	s := client.(*net.TCPConn)
	d := dest.(*net.TCPConn)

	log.Printf("%s: CONNECT %s\n", s.RemoteAddr().String(), host)

	cp := &copier.CancellableCopier{
		LeftConn:     s,
		RightConn:    d,
		ReadTimeout:  10,
		WriteTimeout: 15,
		IOBufsize:    16384,
	}

	cp.Copy(ctx)
}

func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Web Client uses CONNECT method when it sees the target URL is HTTPS and there's a proxy set
	// Non-HTTPS doesn't use this method, whether there's a proxy set or not
	// Details: https://stackoverflow.com/questions/11697943/when-should-one-use-connect-and-get-http-methods-at-http-proxy-server
	if r.Method == "CONNECT" {
		p.handleConnect(w, r)
		return
	}

	if !r.URL.IsAbs() {
		log.Printf("%s: non-proxy req for %q\n", r.Host, r.URL.String())
		http.Error(w, "No support for non-proxy requests", 500)
		return
	}

	t0 := time.Now()

	ctx := r.Context()

	req := r.WithContext(ctx)
	if r.ContentLength == 0 {
		req.Body = nil
	}

	req.Header = cloneCleanHeader(r.Header)
	req.Close = false

	res, err := p.transport.RoundTrip(r)
	if err != nil {
		log.Printf("%s: %s\n", r.Host, err)
		http.Error(w, err.Error(), 500)
		return
	}

	t1 := time.Now()

	copyHeader(w.Header(), res.Header)

	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(res.Trailer)
	if announcedTrailers > 0 {
		trailerKeys := make([]string, 0, len(res.Trailer))
		for k := range res.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		w.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	w.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush()
		}
	}

	nr, err := io.Copy(w, res.Body)
	if err != nil {
		log.Printf("%s: %s\n", r.Host, err)
		http.Error(w, err.Error(), 500)
		return
	}
	res.Body.Close() // close now, instead of defer, to populate res.Trailer

	if len(res.Trailer) == announcedTrailers {
		copyHeader(w.Header(), res.Trailer)
	} else {
		for k, vv := range res.Trailer {
			k = http.TrailerPrefix + k
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
	}

	t2 := time.Now()

	log.Printf("%s: %d %d %s %s\n", r.Host, res.StatusCode, nr, t2.Sub(t0), r.URL.String())

	// Timing log
	d0 := format(t1.Sub(t0))
	d1 := format(t2.Sub(t1))

	now := time.Now().UTC().Format(time.RFC3339)

	log.Printf("time=%q url=%q status=\"%d\" bytes=\"%d\" upstream=%q downstream=%q\n", now, r.URL.String(), res.StatusCode, nr, d0, d1)
}

func (p *HTTPProxy) Close() error {
	return p.lis.Close()
}

func (p *HTTPProxy) Addr() net.Addr {
	return p.lis.Addr()
}
