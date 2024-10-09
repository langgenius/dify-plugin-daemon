FROM golang:1.22-alpine as builder

# copy project
COPY . /app

# set working directory
WORKDIR /app

# build
RUN go build -o /app/main cmd/server/main.go

FROM alpine:latest

COPY --from=builder /app/main /app/main

CMD ["./main"]
