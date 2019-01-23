FROM golang:1.10 as builder
ADD . /go/src/github.com/phylake/hellohttp
WORKDIR /go/src/github.com/phylake/hellohttp
ENV CGO_ENABLED=0
RUN go build -a --installsuffix cgo -ldflags '-s' -o main

FROM centurylink/ca-certs
COPY --from=builder /go/src/github.com/phylake/hellohttp/main /
ENTRYPOINT ["/main"]
