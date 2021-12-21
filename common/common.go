package common

import (
	"fmt"
	"net"
	"strings"
)

// CollectEntry 收集日志的配置的结构体
type CollectEntry struct {
	Path  string `json:"path"`
	Topic string `json:"topic"`
}

// GetOutBoundIp 获取本机ip
func GetOutBoundIp() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	fmt.Println(localAddr.String())
	ip = strings.Split(localAddr.IP.String(), ":")[0]
	return
}
