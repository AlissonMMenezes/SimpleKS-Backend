# Build Container
FROM golang:1.15-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -mod=readonly -v -o server

# Application Container
FROM debian:buster-slim
RUN mkdir /app
COPY simpleks-backend /app/server
CMD ["/app/server"]
