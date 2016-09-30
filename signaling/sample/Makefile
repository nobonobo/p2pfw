PKG=github.com/nobonobo/p2pfw/signaling/sample
export GOPATH:=$(shell cd ../../../../../..; pwd):$(GOPATH)
OPT=-c master -u master

depends:
	go get github.com/goxjs/websocket
	go get golang.org/x/net/websocket

build: depends
	go build $(PKG)

run: build
	./sample $(OPT)
