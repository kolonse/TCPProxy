package TCPProxyProto

const (
	RPOXY_PROTO_SUCCESS      = iota
	RPOXY_PROTO_ERROR_PARAM  = 10001
	RPOXY_PROTO_ERROR_LENGTH // 长度错误
	RPOXY_PROTO_ERROR_FORMAT // 格式错误
	RPOXY_PROTO_ERROR_NET    = 20001
	RPOXY_PROTO_ERROR_LOGIC  = 30001
	RPOXY_PROTO_ERROR_NOT_PROXY_PROTO
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
