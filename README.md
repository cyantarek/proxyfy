# Proxyfy - Go Proxy Server

### Instructions for running the proxy server

01. Run `make build`
02. Modify your proxy config at `config/proxy.conf`
03. Run the proxy using `./proxyfy config/proxy.conf` or `make run`

You should see something like:

`2021/09/25 21:03:27.331418 httpproxy.go:85: Starting HTTP proxy .. at :127.0.0.1:9090`

You can now set this as your proxy server

That's it
