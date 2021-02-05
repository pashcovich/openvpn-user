FROM golang:1.14.2-alpine3.11 AS builder
COPY . /app
RUN cd /app && env CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags='-linkmode external -extldflags "-static" -s -w' -o openvpn-user

