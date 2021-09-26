package proxies

import (
	"log"

	"proxyfy/config"
	"proxyfy/internal/proxies/httpproxy"
	"proxyfy/pkg/privileges"
)

type Proxy interface {
	Start()
	Stop()
}

type ProxyManager struct {
	proxies []Proxy
}

const (
	logFlags = log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds
)

func NewProxyManager(cfg *config.Conf) *ProxyManager {
	log.SetFlags(logFlags)

	var proxyServers []Proxy

	for _, v := range cfg.Http {
		if len(v.Listen) == 0 {
			log.Fatal("http listen address is empty?")
		}
		s, err := httpproxy.New(&v)
		if err != nil {
			log.Fatalf("Can't create http listener on %v: %s", v, err)
		}

		proxyServers = append(proxyServers, s)
	}

	return &ProxyManager{proxies: proxyServers}
}

func (pm *ProxyManager) Run(cfg *config.Conf) {
	// strip down privileges before starting the servers
	privileges.DropPrivilege(cfg.Uid, cfg.Gid)

	for _, s := range pm.proxies {
		s.Start()
	}
}

func (pm ProxyManager) Shutdown() {
	for _, s := range pm.proxies {
		s.Stop()
	}
}
