FROM golang:1.22-alpine as builder

# copy project
COPY . /app

# set working directory
WORKDIR /app

# using goproxy if you have network issues
# ENV GOPROXY=https://goproxy.cn,direct

# build
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main cmd/server/main.go

FROM alpine:latest

COPY --from=builder /app/main /app/main

WORKDIR /app

ENV PLATFORM=aws_lambda
ENV GIN_MODE=release

CMD ["./main"]
