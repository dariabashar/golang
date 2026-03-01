FROM golang:1.25.5-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o myapp ./cmd/api/main.go

FROM alpine:3.19

WORKDIR /

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/myapp /myapp

EXPOSE 8080

ENTRYPOINT ["/myapp"]

