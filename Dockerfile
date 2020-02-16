FROM golang:1.13.7
WORKDIR /webhook

ENV GOPROXY=https://proxy.golang.org
COPY go.mod go.sum /webhook/
RUN go mod download

COPY cmd cmd
RUN CGO_ENABLED=0 GOOS=linux go install -ldflags="-w -s" -v github.com/stevesloka/validatingwebhook/cmd/webhook

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/webhook /bin/webhook