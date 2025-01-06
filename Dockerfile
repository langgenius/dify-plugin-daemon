FROM golang:1.22-alpine as builder

ARG VERSION=unknown

# copy project
COPY . /app

# set working directory
WORKDIR /app

# using goproxy if you have network issues
# ENV GOPROXY=https://goproxy.cn,direct

# build
RUN go build -ldflags "-X 'internal.manifest.VersionX=${VERSION}' -X 'internal.manifest.BuildTimeX=$(date -u +%Y-%m-%dT%H:%M:%S%z)'" -o /app/main cmd/server/main.go

FROM ubuntu:24.04

COPY --from=builder /app/main /app/main

WORKDIR /app

# check build args
ARG PLATFORM=local

# Install python3.12 if PLATFORM is local
RUN if [ "$PLATFORM" = "local" ]; then \
    apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y python3.12 python3.12-venv python3.12-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1; \
    fi

ENV PLATFORM=$PLATFORM
ENV GIN_MODE=release

CMD ["/bin/bash", "-c", "/app/main"]
