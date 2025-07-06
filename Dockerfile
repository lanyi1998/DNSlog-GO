FROM golang:alpine AS builder
WORKDIR /DNSlog-GO
COPY . /DNSlog-GO
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-w -s" -o main cmd/server/main.go

FROM alpine AS runner
WORKDIR /DNSlog-GO
COPY --from=builder /DNSlog-GO/main .
COPY --from=builder /DNSlog-GO/config.yaml .
RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime  \
    && echo Asia/Shanghai > /etc/timezone \
    && apk del tzdata
EXPOSE 53/udp 8000
ENTRYPOINT ["./main"]