package client

import (
	"fmt"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"sync"

	"github.com/goxjs/websocket"
	"github.com/nobonobo/p2pfw/signaling"
)

var (
	noneVal = struct{}{}
	// None ...
	None = &noneVal
	// DefaultSignalingServer ...
	DefaultSignalingServer = "wss://signaling.arukascloud.io/ws"
)

// Config ...
type Config struct {
	signaling.Request
	URL    string
	Origin string
}

// Client ...
type Client struct {
	sync.RWMutex
	*rpc.Client
	dialer func() (*rpc.Client, error)
	config *Config
	closed bool
}

// New ...
func New(config *Config) (*Client, error) {
	if config == nil {
		config = new(Config)
	}
	if len(config.URL) == 0 {
		config.URL = DefaultSignalingServer
	}
	if len(config.UserID) == 0 {
		uuid, err := UUID()
		if err != nil {
			return nil, err
		}
		config.UserID = uuid
	}
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	if len(config.Origin) == 0 {
		origin := *u
		switch origin.Scheme {
		case "wss":
			origin.Scheme = "https"
		case "ws":
			origin.Scheme = "http"
		default:
			return nil, fmt.Errorf("not supported scheme: %s", u.Scheme)
		}
		origin.Path = ""
		config.Origin = origin.String()
	}
	return &Client{
		dialer: func() (*rpc.Client, error) {
			conn, err := websocket.Dial(config.URL, config.Origin)
			if err != nil {
				return nil, err
			}
			return jsonrpc.NewClient(conn), nil
		},
		config: config,
	}, nil
}

// Go ...
func (client *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	client.Lock()
	rpcClient := client.Client
	if rpcClient == nil {
		c, err := client.dialer()
		if err != nil {
			defer client.Unlock()
			call := new(rpc.Call)
			call.ServiceMethod = serviceMethod
			call.Args = args
			call.Reply = reply
			if done == nil {
				call.Done = make(chan *rpc.Call, 10) // buffered.
			} else {
				if cap(done) == 0 {
					log.Panic("rpc: done channel is unbuffered")
				}
			}
			call.Done = done
			call.Error = err
			return call
		}
		client.Client = c
		rpcClient = c
	}
	client.Unlock()
	return rpcClient.Go(serviceMethod, args, reply, done)
}

// Call ...
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) error {
	client.Lock()
	rpcClient := client.Client
	if rpcClient == nil {
		c, err := client.dialer()
		if err != nil {
			defer client.Unlock()
			return err
		}
		client.Client = c
		rpcClient = c
	}
	client.Unlock()
	return rpcClient.Call(serviceMethod, args, reply)
}

// Close ...
func (client *Client) Close() error {
	client.Lock()
	defer client.Unlock()
	c := client.Client
	client.Client = nil
	if c != nil {
		return c.Close()
	}
	return nil
}
