package httpproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/opencoff/go-ratelimit"

	"proxyfy/config"
)

type HTTPProxy struct {
	lis       *net.TCPListener
	conf      *config.ListenConf
	limiter   *ratelimit.RateLimiter
	ctx       context.Context
	cancel    context.CancelFunc
	transport *http.Transport
	server    *http.Server
	wg        sync.WaitGroup
}

func New(lc *config.ListenConf) (*HTTPProxy, error) {
	addr := lc.Listen

	la, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("can't resolve %s: %s", addr, err)
	}

	ln, err := net.ListenTCP("tcp", la)
	if err != nil {
		return nil, fmt.Errorf("can't listen on %s: %s", addr, err)
	}

	rl, err := ratelimit.New(lc.RateLimit.Global, lc.RateLimit.PerHost, 10000)
	if err != nil {
		return nil, fmt.Errorf("%s: can't setup rate limiter: %s", addr, err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	d := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 10 * time.Second,
	}

	p := HTTPProxy{
		lis:     ln,
		conf:    lc,
		limiter: rl,
		ctx:     ctx,
		cancel:  cancel,

		transport: &http.Transport{
			DialContext:         d.DialContext,
			TLSHandshakeTimeout: 8 * time.Second,
			MaxIdleConnsPerHost: 32,
			IdleConnTimeout:     60 * time.Second,
		},

		server: &http.Server{
			Addr:           addr,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		},
	}

	p.server.Handler = &p

	return &p, nil
}

// Start listener
func (p *HTTPProxy) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		log.Printf("Starting HTTP proxy .. at :%s\n", p.lis.Addr())
		p.server.Serve(p)
	}()
}

// Stop server
func (p *HTTPProxy) Stop() {
	p.cancel()
	p.lis.Close() // causes Accept() to abort

	cx, cancel := context.WithTimeout(p.ctx, 10*time.Second)
	p.server.Shutdown(cx)
	cancel()

	p.wg.Wait()
	log.Printf("HTTP proxy shutdown\n")
}
