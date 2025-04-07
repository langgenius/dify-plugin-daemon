FROM golang:1.23-alpine as builder

ARG VERSION=unknown

# copy project
COPY . /app

# set working directory
WORKDIR /app

# using goproxy if you have network issues
# ENV GOPROXY=https://goproxy.cn,direct

# build
RUN CGO_ENABLED=0 go build \
    -ldflags "\
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.VersionX=${VERSION}' \
    -X 'github.com/langgenius/dify-plugin-daemon/internal/manifest.BuildTimeX=$(date -u +%Y-%m-%dT%H:%M:%S%z)'" \
    -o /app/main cmd/server/main.go

FROM alpine:latest

COPY --from=builder /app/main /app/main

WORKDIR /app

# check build args
ARG PLATFORM=serverless

ENV PLATFORM=$PLATFORM
ENV GIN_MODE=release

# run the server
CMD ["./main"]
