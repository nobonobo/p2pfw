NAME := turnserver
TAG  := nobonobo/$(NAME)
PKG=github.com/nobonobo/p2pfw/turnserver
export GOPATH:=$(shell cd ../../../../..; pwd):$(GOPATH)
TARVER := 4.5.0.4
TARBALL := turnserver-$(TARVER).tar.gz

push: build
	docker push $(TAG)

depends:
	go get github.com/rs/cors

$(TARBALL):
	curl http://coturn.net/turnserver/v$(TARVER)/$(TARBALL) -o $(TARBALL)

build: depends $(TARBALL)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(PKG)
	docker build --rm -t $(TAG) .

run:
	docker run -it --rm --name $(NAME) -p 8080:8080 $(TAG)

shell:
	docker exec -it $(NAME) ash
