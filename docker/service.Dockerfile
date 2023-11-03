# TODO: Replace 'servicename' with the approriate name for your service. This is just the folder name within
# the container but you should rename them all to make debugging easier.

FROM golang:1.20.3-alpine3.17 as builder

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum

COPY pkg/ pkg/
COPY cmd/ cmd/

# TODO: Replace 'servicename'
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /build/servicename ./cmd/servicename

FROM alpine:3.17.3

WORKDIR /app

# TODO: Replace 'servicename'
COPY --from=builder /build/servicename .

# TODO: Replace 'servicename'
CMD ["/app/servicename"]
