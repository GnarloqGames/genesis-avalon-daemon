#!/bin/bash
ver=$(git describe --tags --abbrev)

for arch in "amd64" "arm64"
do
    docker push registry.0x42.in/library/docker/genesis-avalon-avalond:${ver}-${arch}
done

docker buildx imagetools create \
    -t registry.0x42.in/library/docker/genesis-avalon-avalond:${ver} \
    registry.0x42.in/library/docker/genesis-avalon-avalond:${ver}-amd64 \
    registry.0x42.in/library/docker/genesis-avalon-avalond:${ver}-arm64