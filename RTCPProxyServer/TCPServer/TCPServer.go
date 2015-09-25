// TCPServer
package TCPServer

import (
	"KolonseWeb"
	. "TCPProxyProto"
	"fmt"
	//	"io"
	"net"
	"regexp"
	"strconv"
	"time"
)

const (
	SERVER_STATUS_NOTEXIST       = "not exist"
	SERVER_STATUS_WAIT           = "wait"
	SERVER_STATUS_RUNNING        = "running"
	SERVER_STATUS_TCP_ADDR_ERROR = "tcp addr error"
	SERVER_STATUS_LISTEN_FAIL    = "listen fail"
)

type Who struct {
	Name string       // 服务者名字
	Ip   string       // 客户端 IP
	conn *net.TCPConn // 连接套接字
}

func (w *Who) Close() {
	if w.conn != nil {
		(*w.conn).Close()
	}
}

func (w *Who) Dump() string {
	ret := ""
	ret += fmt.Sprintln("\tName:", w.Name)
	ret += fmt.Sprintln("\tIp:", w.Ip)
	return ret
}

func (w *Who) Set(name string, ip string, conn *net.TCPConn) {
	w.Close()
	w.conn = conn
	w.Name = name
	w.Ip = ip
}

type ServerInfo struct {
	Domain string
	Port   uint16
	Desc   string
	Status string // 服务状态
	ForWho Who
	Parter Who
}

func (si *ServerInfo) GetStatus() string {
	return si.Status
}

func (si *ServerInfo) Dump() string {
	ret := ""
	ret += fmt.Sprintln("Domain:", si.Domain)
	ret += fmt.Sprintln("Port:", si.Port)
	ret += fmt.Sprintln("Desc:", si.Desc)
	ret += fmt.Sprintln("Status:", si.Status)
	ret += fmt.Sprintln("ForWho:")
	ret += si.ForWho.Dump()
	return ret
}

func (si *ServerInfo) ReadLessNByte(conn net.Conn, nBytes int, buff []byte) (int, error) {
	totalRecv := 0
	for {
		// 设置超时时间为10S  如果没有读取到足够的数据 那么表示非客户端
		conn.SetReadDeadline(time.Now().Add(time.Second * 10)) // 10S后超时
		n, err := conn.Read(buff)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			return totalRecv + n, err
		}
		totalRecv += n
		if totalRecv >= nBytes {
			return totalRecv, nil
		}
	}
}

func (si *ServerInfo) handleConnection(conn *net.TCPConn) {
	KolonseWeb.DefaultLogs().Info("Recv Conn,RemoteAddr:%v %v %v %v",
		conn.RemoteAddr().Network(), conn.RemoteAddr().String(),
		conn.LocalAddr().Network(), conn.LocalAddr().String())
	// 收到一个连接 读取开始 i'm proxy server
	buff := make([]byte, 10000) //  缓存长度
	n, err := si.ReadLessNByte(conn, len(PROXY_PROTO_HEAD_MARK), buff)
	// 需要检测 time out
	bClose := false
	var validID = regexp.MustCompile(`.+i/o timeout`)
	if validID.MatchString(err.Error()) {
		KolonseWeb.DefaultLogs().Warning("Addr r_%v|l_%v Warning:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
	} else if err != nil {
		KolonseWeb.DefaultLogs().Error("Addr r_%v|l_%v Error:%v", conn.RemoteAddr(), conn.LocalAddr(), err.Error())
		conn.Close()
		bClose = true
	}
	// 检测是否是 PROXY SERVER
	if CheckHeadMark(buff) {
		if !bClose {
			si.ProcessProxyConn(conn, n, buff)
		}
	} else { // 如果是请求连接 那么需要 进行连接会话处理
		if !bClose { // 如果关闭连接 那么就需要通知关闭
			si.ProcessReqConn(conn, n, buff)
		}
	}
}

func (si *ServerInfo) ProcessProxyConn(conn *net.TCPConn, nRecvdBytes int, buff []byte) {

}

func (si *ServerInfo) ProcessReqConn(conn *net.TCPConn, nRecvdBytes int, buff []byte) {

	return
}

func (si *ServerInfo) Start() error {
	host := ":" + strconv.Itoa(int(si.Port))
	tcpAddr, err := net.ResolveTCPAddr("tcp4", host)
	if err != nil {
		si.Status = SERVER_STATUS_TCP_ADDR_ERROR
		return err
	}
	ln, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		si.Status = SERVER_STATUS_LISTEN_FAIL
		return err
	}
	si.Status = SERVER_STATUS_RUNNING
	KolonseWeb.DefaultLogs().Info("TCP Server Listen On %v", si.Port)
	go func() {
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				time.Sleep(time.Millisecond * 100)
				continue
			}
			go si.handleConnection(conn)
		}
	}()
	return nil
}

func NULLServerInfo() *ServerInfo {
	return &ServerInfo{
		Status: SERVER_STATUS_NOTEXIST,
	}
}

func NewServerInfo() *ServerInfo {
	return &ServerInfo{
		Status: SERVER_STATUS_WAIT,
	}
}
