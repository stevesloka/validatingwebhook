FROM golang:1.10.1
WORKDIR /go/src/github.com/stevesloka/webhook

COPY cmd cmd
COPY vendor vendor
RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/stevesloka/webhook/cmd/webhook

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/webhook /bin/webhook