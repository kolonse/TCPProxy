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

func Test_ParseBody(t *testing.T) {
	pp := NewProxyProto()
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_HEAD_MARK)...)
	pp.buff = append(pp.buff, []byte(PROXY_PROTO_BODY_LENGTH+"10\r\n")...)
	pp.buff = append(pp.buff, []byte(RPOXY_PROTO_HEAD_END)...)
	pp.buff = append(pp.buff, []byte("0123456789")...)
	err := pp.ParseMark().ParseBody().GetError()
	if err.GetCode() != 0 {
		t.Error(err.Error())
	}
	if string(pp.GetBody()) != "0123456789" {
		t.Error("body error")
	}

	fmt.Println(string(pp.GetBody()))
}

func Test_Stringify(t *testing.T) {
	pp := NewProxyProto()
	pp.StringifyConn().
		StringifyLocalAddr("12345").
		StringifyRemoteAddr("56789").
		StringifyBody([]byte("0123456789")).
		StringifyEnd()
	fmt.Println(string(pp.GetBuff()))
	pp_t := NewProxyProto()
	err := pp_t.Parse(pp.GetBuff()).GetError()
	fmt.Println(string(pp_t.GetBuff()))
	if err.GetCode() != 0 {
		t.Error(err.Error())
		return
	}

	if string(pp_t.GetBody()) != "0123456789" {
		t.Error(string(pp_t.GetBody()))
		return
	}
	//if string(pp_t.GetLocalAddr()) != "12345" {
	//	t.Error("GetLocalAddr error")
	//	return
	//}

	//if string(pp_t.GetRemoteAddr()) != "56789" {
	//	t.Error("GetRemoteAddr error")
	//	return
	//}

	if string(pp_t.GetMethod()) != "CONN" {
		t.Error(pp_t.GetMethod())
		return
	}
}
