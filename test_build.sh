#!/bin/bash

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X 'github.com/nextdotid/proof_server/common.Environment=production' -X 'github.com/nextdotid/proof_server/common.BuildTime=$(date +%s)'" -o headless_amd64 ./cmd/headless
docker build -t ghcr.io/nextdotid/proof_service_headless:latest --build-arg TARGETARCH=amd64 -f .github/workflows/docker/Dockerfile.headless .
# docker run --rm -it -p 9801:9801 ghcr.io/nextdotid/proof_service_headless:latest
docker save ghcr.io/nextdotid/proof_service_headless:latest | ssh oracle_primary docker load
