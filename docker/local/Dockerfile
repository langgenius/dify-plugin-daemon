FROM golang:1.22-alpine as builder

# copy project
COPY . /app

# set working directory
WORKDIR /app

# using goproxy if you have network issues
# ENV GOPROXY=https://goproxy.cn,direct

# build
RUN go build -o /app/main cmd/server/main.go

FROM ubuntu:22.04

COPY --from=builder /app/main /app/main

WORKDIR /app

# Install python3.10
RUN apt-get update && apt-get install -y python3.10 python3.10-venv python3.10-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV PLATFORM=local
ENV GIN_MODE=release

CMD ["/app/main"]
