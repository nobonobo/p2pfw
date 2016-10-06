package peerconn

import (
	"io"

	"github.com/nobonobo/webrtc"
)

type connWrapper struct {
	*webrtc.DataChannel
	dst io.Reader
}

// NewDCConn ...
func NewDCConn(channel *webrtc.DataChannel) io.ReadWriteCloser {
	dst, src := io.Pipe()
	channel.OnMessage(func(b []byte) {
		src.Write(b)
	})
	return &connWrapper{
		DataChannel: channel,
		dst:         dst,
	}
}

func (w *connWrapper) Read(b []byte) (int, error) {
	return w.dst.Read(b)
}

func (w *connWrapper) Write(b []byte) (int, error) {
	w.DataChannel.Send(b)
	return len(b), nil
}
