package TCPProxyProto

import (
	"fmt"
	"testing"
)

func Test_ParseMark(t *testing.T) {
	pp := NewProxyProto()
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_HEAD_MARK+RPOXY_PROTO_HEAD_END)...)
	err := pp.ParseMark().GetError()
	if err.GetCode() != 0 {
		t.Error(err.Error())
	}
	fmt.Println(pp.HeaderString())
}

func Test_ParseBodyLength(t *testing.T) {
	pp := NewProxyProto()
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_HEAD_MARK)...)
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_BODY_LENGTH+"100\r\n")...)
	pp.buff = append(pp.buff, []byte(RPOXY_PROTO_HEAD_END)...)
	err := pp.ParseMark().ParseBodyLength().GetError()
	if err.GetCode() != 0 {
		t.Error(err.Error())
	}
	fmt.Println(pp.HeaderString())
}

func Test_ParseBody(t *testing.T) {
	pp := NewProxyProto()
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_HEAD_MARK)...)
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_BODY_LENGTH+"10\r\n")...)
	pp.buff = append(pp.buff, []byte(RPOXY_PROTO_HEAD_END)...)
	pp.buff = append(pp.buff, []byte("0123456789")...)
	err := pp.ParseMark().ParseBodyLength().ParseBody().GetError()
	if err.GetCode() != 0 {
		t.Error(err.Error())
	}
	if string(pp.GetBody()) != "0123456789" {
		t.Error("body error")
	}

	fmt.Println(string(pp.GetBody()))
}
