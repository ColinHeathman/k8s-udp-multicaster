# Build stage
FROM golang:1.10.1 AS build

WORKDIR /go/src/unbc.ca/net-tool

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Final run stage
FROM scratch AS run

COPY --from=build  /go/src/unbc.ca/net-tool/app .

ENV PROTOCOL="udp"
ENV LISTEN_HOST=""
ENV DIAL_HOST="127.0.0.0"
ENV PORT="9782"
ENV TYPE="LISTEN"
ENV POD_NAME="pod_name"
ENV POD_IP="0.0.0.0"
ENV NODE_NAME="node_name"

ENTRYPOINT ["./app"]
