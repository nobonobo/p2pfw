NAME := signaling
TAG  := nobonobo/$(NAME)
PKG=github.com/nobonobo/p2pfw/signaling/server
export GOPATH:=$(GOPATH):$(shell cd ../../../../../..; pwd)

push: build
	docker push $(TAG)

build:
	docker build --rm -t $(TAG) .

run:
	docker run -it --rm --name $(NAME) -p 8080:8080 $(TAG)

shell:
	docker exec -it $(NAME) ash
