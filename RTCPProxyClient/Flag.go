// Flag
package main

import (
	"flag"
)

var ProxyDomain = flag.String("d", "pay.wanchuangyou.com", "需要逆向代理的域名")
var ProxyServer = flag.String("ps", "127.0.0.1:7020", "代理映射到服务的地址")
var ProxyName = flag.String("pn", "kolonse", "代理名字")
var UsePort = flag.Int("up", 9988, "代理服务使用的端口")
var Server = flag.String("server", "127.0.0.1:9877", "逆向代理服务 HTTP Server 地址")

func init() {
	flag.Parse()
}
