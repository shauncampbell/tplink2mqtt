FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/github.com/shauncampbell/tplink2mqtt/
COPY . .

RUN go build -o /go/bin/tplink2mqtt ./cmd/tplink2mqtt

FROM alpine:3.12

COPY --from=builder /go/bin/tplink2mqtt /go/bin/tplink2mqtt
LABEL maintainer="Shaun Campbell <docker@shaun.scot>"

ENV TPLINK_MQTT_HOST ""
ENV TPLINK_MQTT_PORT 1883
ENV TPLINK_MQTT_USERNAME ""
ENV TPLINK_MQTT_PASSWORD ""
ENV TPLINK_SUBNET "192.168.0.2/24"
ENV TPLINK_TIMEOUT 5
ENV TPLINK_INTERVAL 60

ENTRYPOINT ["./go/bin/tplink2mqtt"]