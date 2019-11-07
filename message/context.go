package message

import "net/http"

type Context struct {
	Data     []byte
	DataType int
	Req      *http.Request
}
