FROM golang:1.16-alpine AS builder

ENV GO111MODULE=on

#WORKDIR /go/src/github.com/gusaul/grpcox
WORKDIR /src

COPY . ./
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o grpcox grpcox.go


FROM alpine

COPY --from=builder /src/grpcox ./
RUN mkdir /log
EXPOSE 6969
ENTRYPOINT ["./grpcox", "-log", "log/grpcox.log"]
