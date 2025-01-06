FROM golang:1.22-alpine as builder

# copy project
COPY . /app

# set working directory
WORKDIR /app

# using goproxy if you have network issues
# ENV GOPROXY=https://goproxy.cn,direct

# build
RUN go build -o /app/main cmd/server/main.go

FROM ubuntu:24.04

COPY --from=builder /app/main /app/main

WORKDIR /app

# Install python3.12
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y python3.12 python3.12-venv python3.12-dev \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.12 1

ENV PLATFORM=local
ENV GIN_MODE=release

CMD ["/app/main"]
