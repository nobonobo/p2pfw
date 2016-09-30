package jsonrpc

import (
	"io"
	"net/http"
	"net/rpc"
	org "net/rpc/jsonrpc"
)

type scodec struct {
	ch chan<- error
	rpc.ServerCodec
}

// WriteResponse ...
func (c *scodec) WriteResponse(r *rpc.Response, x interface{}) error {
	err := c.ServerCodec.WriteResponse(r, x)
	c.ch <- c.ServerCodec.Close()
	return err
}

// Handle ...
var Handle = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		ch := make(chan error, 1)
		conn := &struct {
			io.ReadCloser
			io.Writer
		}{r.Body, w}
		rpc.ServeCodec(&scodec{ch, org.NewServerCodec(conn)})
		<-ch
	default:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 POST only\n")
	}
})
