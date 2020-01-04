# build stage
FROM golang:1.13 AS builder

WORKDIR /go/src/github.com/ko-da-k/github-developer-exporter
COPY . .

RUN make build-linux

# final stage
FROM alpine:3.10

RUN apk add --no-cache ca-certificates && \
    addgroup -S appusers && adduser -S -G appusers appuser

COPY --from=builder /go/src/github.com/ko-da-k/github-developer-exporter/app /app/

WORKDIR /app

# please change timezone
RUN set -o xtrace && \
    chmod 755 ./app && \
    : Set timezone to JST && \
        apk add --no-cache --virtual .tz tzdata && \
        cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
        apk del --purge .tz

USER appuser

ENTRYPOINT [ "./app" ]
