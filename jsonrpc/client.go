package jsonrpc

import (
	"io"
	"net/http"
	"net/rpc"
	org "net/rpc/jsonrpc"
)

type ccodec struct {
	c io.Closer
	rpc.ClientCodec
}

func (c *ccodec) WriteRequest(r *rpc.Request, x interface{}) error {
	err := c.ClientCodec.WriteRequest(r, x)
	c.c.Close()
	return err
}

// Dial ...
func Dial(endpoint string) *rpc.Client {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()
	go func() {
		resp, err := http.Post(endpoint, "application/json", r1)
		if err != nil {
			w2.CloseWithError(err)
			return
		}
		defer resp.Body.Close()
		_, err = io.Copy(w2, resp.Body)
		if err != nil {
			w2.CloseWithError(err)
		}
	}()
	return rpc.NewClientWithCodec(&ccodec{
		c: w1,
		ClientCodec: org.NewClientCodec(&struct {
			io.Reader
			io.Writer
			io.Closer
		}{
			r2,
			w1, //io.MultiWriter(w1, os.Stdout),
			w2,
		}),
	})
}

// Client ...
type Client struct {
	endpoint string
	dialer   func(endoint string) *rpc.Client
}

// NewClient ...
func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
		dialer:   Dial,
	}
}

// Go ...
func (c *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	client := c.dialer(c.endpoint)
	return client.Go(serviceMethod, args, reply, done)
}

// Call ...
func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	client := c.dialer(c.endpoint)
	return client.Call(serviceMethod, args, reply)
}

// Close ...
func (c *Client) Close() error {
	return nil
}
