package proxy

import "encoding/json"

type JSONRpcReq struct {
	Id     *json.RawMessage `json:"id"`
	Method string           `json:"method"`
	Params *json.RawMessage `json:"params"`
}

type JSONRpcResp struct {
	Id      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result"`
	Error   interface{}      `json:"error,omitempty"`
}

type JobReplyData struct {
	Data   string `json:"data"`
	Target string `json:"target"`
}

type SubmitReply struct {
	Status string `json:"status"`
}

type ErrorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
