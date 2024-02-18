FROM registry.0x42.in/library/docker/genesis-avalon-builder:bookworm-1.1.0 as builder

WORKDIR /build
COPY . .
RUN go build -o ./bin/avalond ./cmd/daemon/...

FROM debian:bookworm
COPY --from=builder /build/bin/avalond /usr/bin/avalond