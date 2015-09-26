// Flag
package main

import (
	"flag"
)

var ProxyDomain = flag.String("domain", "pay.wanchuangyou.com", "proxy domain")
var ProxyServer = flag.String("pserver", "127.0.0.1:12345", "proxy server")
var Server = flag.String("server", "127.0.0.1:9877", "server addr")
var ProxyName = flag.String("name", "kolonse", "proxy name")

func init() {
	flag.Parse()
}
