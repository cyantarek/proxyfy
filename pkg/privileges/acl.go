package privileges

import (
	"net"

	"proxyfy/config"
)

// ACLCheck Returns true if the new connection 'conn' passes the ACL checks
// Return false otherwise
func ACLCheck(cfg *config.ListenConf, conn net.Conn) bool {
	h, ok := conn.RemoteAddr().(*net.TCPAddr)
	if !ok {
		//p.log.Debug("%s can't extract TCP Addr", conn.RemoteAddr().String())
		return false
	}

	for _, n := range cfg.Deny {
		if n.Contains(h.IP) {
			return false
		}
	}

	if len(cfg.Allow) == 0 {
		return true
	}

	for _, n := range cfg.Allow {
		if n.Contains(h.IP) {
			return true
		}
	}

	return false
}
