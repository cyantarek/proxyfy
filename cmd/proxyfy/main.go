package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"proxyfy/config"
	"proxyfy/internal/proxies"
)

func main() {
	usage := fmt.Sprintf("%s config-file", "proxy")

	// any files we create will be readable ONLY by us
	syscall.Umask(0077)

	args := os.Args
	if len(args) < 2 {
		log.Fatalf("No config file!\nUsage: %s", usage)
	}

	cfg, err := config.ReadYAML(args[1])
	if err != nil {
		log.Fatalf("Can't read config file %s: %s", args[0], err)
	}

	pm := proxies.NewProxyManager(cfg)

	go pm.Run(cfg)

	// Setup signal handlers
	sigChan := make(chan os.Signal, 4)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	signal.Ignore(syscall.SIGPIPE, syscall.SIGFPE)

	// Now wait for signals to arrive
	for {
		s := <-sigChan
		t := s.(syscall.Signal)

		log.Printf("Caught signal %d; Terminating ..\n", int(t))
		break
	}

	pm.Shutdown()
}
