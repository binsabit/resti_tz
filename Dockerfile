FROM golang:1.21.10-alpine3.19 AS builder

WORKDIR /app

RUN apk add --no-cache tzdata

COPY go.mod .
COPY go.sum .

RUN go mod download && go mod verify

COPY . .

RUN  go build -v -ldflags="-w -s" -o main ./cmd/main.go

FROM alpine:3.18
WORKDIR /app

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

ENV TZ=Asia/Almaty

COPY --from=builder /app .

CMD ["sh","-c","./main"]
