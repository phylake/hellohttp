FROM golang:1.13 as builder
ADD . /go/src/github.com/phylake/hellohttp
WORKDIR /go/src/github.com/phylake/hellohttp
CMD go run main.go
