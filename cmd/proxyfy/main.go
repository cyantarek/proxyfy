package main

import (
	"os"
	"os/signal"
	"syscall"

	"proxyfy/config"
	"proxyfy/internal/proxies"
	"proxyfy/pkg/logger"
)

func main() {
	// any files we create will be readable ONLY by us
	syscall.Umask(0077)

	var configFile string

	if len(os.Args) < 2 {
		logger.Log.Warningln("no config file provided. Using the default config!")

		configFile = "config/default.conf"
	}

	cfg, err := config.ReadYAML(configFile)
	if err != nil {
		logger.Log.Fatalf("Can't read config file %s: %s", configFile, err)
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

		logger.Log.Infoln("Caught signal %d; Terminating ..", int(t))
		break
	}

	pm.Shutdown()
}
