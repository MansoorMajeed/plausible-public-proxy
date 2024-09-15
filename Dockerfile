# Stage 1: Build the Go binary
FROM golang:1.23 AS builder


WORKDIR /app

ENV CGO_ENABLED=0
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o plausible-public-proxy .


FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/plausible-public-proxy /app/plausible-public-proxy


EXPOSE 7000

CMD ["/app/plausible-public-proxy"]