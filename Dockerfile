# Build
FROM golang:1 as build

WORKDIR /build
ADD . .

# Tests
RUN go test -mod=vendor ./...

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o actor .

# Run
FROM centos

RUN adduser -s /bin/false actor

COPY actor.sample.yaml /etc/bmc-toolbox/actor.yaml

COPY --from=build /build/actor /usr/bin

EXPOSE 8000
USER actor

ENTRYPOINT ["/usr/bin/actor"]
