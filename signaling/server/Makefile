NAME := signaling
TAG  := nobonobo/$(NAME)
PKG=github.com/nobonobo/p2pfw/signaling/server
export GOPATH:=$(GOPATH):$(shell cd ../../../../../..; pwd)

push: build
	docker push $(TAG)

depends:
	go get github.com/rs/cors
	go get golang.org/x/net/websocket

build: depends
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(PKG)
	docker build --rm -t $(TAG) .

run: build
	docker run -it --rm --name $(NAME) -p 8080:8080 $(TAG)

shell:
	docker exec -it $(NAME) ash
