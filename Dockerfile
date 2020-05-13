FROM golang:latest

WORKDIR /go/src/cordless

COPY . ./

RUN go build .

VOLUME ["/root/.config/cordless"]

ENTRYPOINT ["/go/src/cordless/cordless"]

