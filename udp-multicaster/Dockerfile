# Build stage
FROM golang:1.10.1 AS build

WORKDIR /go/src/unbc.ca/udp-multicaster

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final run stage
FROM scratch AS run

COPY --from=build  /go/src/unbc.ca/udp-multicaster/app .

ENV LISTEN_HOST=""
ENV LISTEN_PORT="9782"
ENV DIAL_HOST="127.0.0.0"
ENV DIAL_PORT="9782"

ENTRYPOINT ["./app"]
