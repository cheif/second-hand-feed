FROM golang:alpine AS builder
WORKDIR $GOPATH/src/second-hand-rss
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
            go build -o /go/bin/second-hand-rss

FROM alpine:latest
COPY --from=builder /go/bin/second-hand-rss /go/bin/second-hand-rss
COPY templates templates
EXPOSE 8080
VOLUME /config
CMD ["/bin/sh", "-c", "/go/bin/second-hand-rss /config/config.json"]
