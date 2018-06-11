
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean

DOCKERCMD=docker

DOCKERBUILD=$(DOCKERCMD) build -t izanagi1995/mikrodns .
DOCKERPUSH=$(DOCKERCMD) push izanagi1995/mikrodns

BINARY=mikrodns

all: build docker docker-push
build:
	go build -a -tags netgo -ldflags '-w' -o $(BINARY) main.go
clean:
	$(GOCLEAN)
	rm -rf $(BINARY)
docker:
	$(DOCKERBUILD)
docker-push:
	$(DOCKERPUSH)
