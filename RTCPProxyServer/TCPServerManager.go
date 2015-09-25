// TCPServerManager
package main

import (
	. "TCPProxy/RTCPProxyServer/TCPServer"
	"encoding/json"
	"os"
)

type TypeServerManager map[string]*ServerInfo

var TCPServerManager TypeServerManager

func (tm *TypeServerManager) GetServerInfo(domain string) *ServerInfo {
	for _, value := range *tm {
		if value.Domain == domain {
			return value
		}
	}
	return NULLServerInfo()
}

func (tm *TypeServerManager) Dump() string {
	ret := ""
	for _, value := range *tm {
		ret += "+++++++++++++++++++++++++++++++++\n"
		ret += value.Dump()
		ret += "=================================\n"
	}
	return ret
}

func (tm *TypeServerManager) TCPServerManagerStart() {
	for _, value := range *tm {
		go value.Start()
	}
}

func LoadCfg() {
	file, err := os.Open("./cfg.rtps")
	if err != nil {
		panic(err)
		return
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&TCPServerManager)
	if err != nil {
		panic(err)
		return
	}
}

func init() {
	TCPServerManager = make(TypeServerManager)
}
