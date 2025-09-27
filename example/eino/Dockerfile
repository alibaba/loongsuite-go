FROM golang:1.23 AS builder

WORKDIR /build

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/main ./cmd/einoagent/*.go

FROM alpine:3.21

WORKDIR /

RUN apk --no-cache add ca-certificates redis \
      && update-ca-certificates

COPY .env /.env
COPY data /data
COPY --from=builder /out/main /main

EXPOSE 8080

CMD [ "/main" ]
