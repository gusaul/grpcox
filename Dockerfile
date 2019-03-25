FROM golang:1.10-alpine AS builder

WORKDIR /go/src/github.com/gusaul/grpcox

COPY . ./
RUN go build -o grpcox grpcox.go


FROM alpine

COPY ./index /index
COPY --from=builder /go/src/github.com/gusaul/grpcox/grpcox ./
EXPOSE 6969
ENTRYPOINT ["./grpcox"]