FROM golang:1.21.5 AS builder

ENV GO111MODULE=on \
   CGO_ENABLED=0 \
   GOOS=linux \
   GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -v -o main .

FROM debian:stretch-slim

COPY --from=builder /app/main /usr/local/bin/

CMD ["main"]