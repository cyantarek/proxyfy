build: build-proxyfy

build-proxyfy:
	go build ./cmd/proxyfy

run:
	./proxyfy config/proxy.conf
