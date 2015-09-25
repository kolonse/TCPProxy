// Flag
package main

import (
	"flag"
)

var ProxyDomain = flag.String("domain", "pay.wanchuangyou.com", "proxy domain")
var Server = flag.String("server", "127.0.0.1:9877", "server addr")
var ProxyName = flag.String("name", "kolonse", "proxy name")

func init() {
	flag.Parse()
}
