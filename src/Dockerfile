# Build stage
FROM golang:1.10.1 AS build

WORKDIR /go/src/unbc.ca/udp-multicaster

RUN go get k8s.io/client-go/...

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go test -v
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final run stage
FROM scratch AS run

COPY --from=build  /go/src/unbc.ca/udp-multicaster/app .

ENV LISTEN_HOST=""
ENV LISTEN_PORT="9782"
ENV SERVICE_NAME=""
ENV SERVICE_PORT="udp"
ENV SYNC_INTERVAL="5s"
ENV UDP_QUEUE_MAX="10"

ENTRYPOINT ["./app"]
