FROM golang:1.14.14-alpine

WORKDIR /go/src/app
COPY . .

ARG GOPROXY="https://goproxy.io,direct"

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["woodpecker"]
