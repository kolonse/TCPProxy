// TCPProxyProto project TCPProxyCommon.go
package TCPProxyProto

import (
	"errors"
	"strconv"
	"strings"
	//	"io"
)

var (
	VERSION = "PROXY/0.1"
)

/**
*	TCP 协议
 */
const (
	PROXY_PROTO_HEAD_MARK        = "-------kolonse rtcp proxy head mark begin-------\r\n"
	PROXY_PROTO_HEAD_PROTO_CONN  = "CONN PROXY/0.1"
	PROXY_PROTO_HEAD_PROTO_REQ   = "REQ PROXY/0.1"
	PROXY_PROTO_HEAD_PROTO_RES   = "RES PROXY/0.1"
	PROXY_PROTO_HEAD_PROTO_CLOSE = "CLOSE PROXY/0.1"
	PROXY_PROTO_BODY_LENGTH      = "Content Length:"
	RPOXY_PROTO_LOCAL_ADDR       = "Local Addr:"
	RPOXY_PROTO_REMOTE_ADDR      = "Remote Addr:"
	RPOXY_PROTO_LINE_END         = "\r\n"
	RPOXY_PROTO_HEAD_END         = "-------kolonse rtcp proxy head mark end-------\r\n"
)

type ProxyProto struct {
	mark       string
	method     string
	version    string
	localAddr  string
	remote     string
	bodyLenght int
	bodyBuff   []byte
	headBuff   []byte
	buff       []byte
	err        error
}

func (pp *ProxyProto) HeaderString() string {
	return string(pp.headBuff)
}

func (pp *ProxyProto) Parse(buff []byte) *ProxyProto {
	pp.buff = make([]byte, len(buff))
	copy(pp.buff, buff)
	return pp.ParseMark().
		ParseBodyLength().
		ParseBody().
		ParseProto().
		ParseVersion().
		ParseLocalAddr().
		ParseRemoteAddr()
}

func (pp *ProxyProto) ParseMark() *ProxyProto {
	if len(pp.buff) < len(PROXY_PROTO_HEAD_MARK) { // buff 长度不够处理 mark 头标志
		pp.err = errors.New("parse mark: length not enougth")
	} else {
		pp.mark = string(PROXY_PROTO_HEAD_MARK)
		// 将 header buff 单独拿出来
		index := strings.Index(string(pp.buff), string(RPOXY_PROTO_HEAD_END))
		if index == -1 {
			pp.err = errors.New("not found head end")
		} else {
			pp.headBuff = make([]byte, index+len(RPOXY_PROTO_HEAD_END))
			copy(pp.headBuff, pp.buff[:index+len(RPOXY_PROTO_HEAD_END)])
		}
	}
	//	if( )
	return pp
}

func (pp *ProxyProto) ParseProto() *ProxyProto { //  协议必定在 mark 之后
	if pp.GetError() == nil {

	}
	return pp
}

func (pp *ProxyProto) ParseLocalAddr() *ProxyProto {
	if pp.GetError() == nil {

	}
	return pp
}
func (pp *ProxyProto) ParseRemoteAddr() *ProxyProto {
	if pp.GetError() == nil {

	}
	return pp
}

func (pp *ProxyProto) ParseVersion() *ProxyProto {
	if pp.GetError() == nil {

	}
	return pp
}

func (pp *ProxyProto) ParseBodyLength() *ProxyProto {
	if pp.GetError() == nil {
		//  查找 body length 字段
		index := strings.Index(string(pp.headBuff), string(PROXY_PROTO_BODY_LENGTH))
		if index == -1 {
			pp.err = errors.New("not found Content Length")
			return pp
		}

		indexEnd := strings.Index(string(pp.headBuff[index:]), string(RPOXY_PROTO_LINE_END))
		if indexEnd == -1 {
			pp.err = errors.New("not found Content Length's Line End")
			return pp
		}

		lengthString := string(pp.headBuff[index+len(PROXY_PROTO_BODY_LENGTH) : indexEnd])
		pp.bodyLenght, pp.err = strconv.Atoi(lengthString)
	}
	return pp
}

func (pp *ProxyProto) ParseBody() *ProxyProto {
	if pp.GetError() == nil {
		if len(pp.buff) < pp.bodyLenght+len(pp.headBuff) { // 长度不足
			pp.err = errors.New("parse body: length not enougth")
			return pp
		}

		pp.bodyBuff = make([]byte, pp.bodyLenght)
		copy(pp.bodyBuff, pp.buff[len(pp.headBuff):])
	}
	return pp
}

func (pp *ProxyProto) GetError() error {
	return pp.err
}

func NewProxyProto() *ProxyProto {
	return &ProxyProto{}
}

type ProtoBuff []byte

func (pb *ProtoBuff) ProxyClose(localAddr, remoteAddr string) {
	length := 0
	*pb = append((*pb)[length:], []byte(PROXY_PROTO_HEAD_MARK)...)
	length += len(PROXY_PROTO_HEAD_MARK)
	*pb = append((*pb)[length:], []byte(PROXY_PROTO_HEAD_PROTO_CLOSE)...)
}

func CheckHeadMark(buff []byte) bool {
	if len(buff) >= len(PROXY_PROTO_HEAD_MARK) && string(buff[0:len(PROXY_PROTO_HEAD_MARK)]) == PROXY_PROTO_HEAD_MARK {
		return true
	}
	return false
}

func NewProtoBuff() *ProtoBuff {
	return &ProtoBuff{}
}

/**
*	HTTP 响应 数据协议
 */
const (
	RPOXY_PROTO_SUCCESS     = iota
	RPOXY_PROTO_ERROR_PARAM = 10001
	RPOXY_PROTO_ERROR_NET   = 20001
	RPOXY_PROTO_ERROR_LOGIC = 20001
)

type RespProto struct {
	Code    int
	Message string
	Extern  interface{}
}

func NewRespProto(code int, message string, extern interface{}) *RespProto {
	return &RespProto{
		Code:    code,
		Message: message,
		Extern:  extern,
	}
}
