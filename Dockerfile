# Build Container
FROM golang:1.15-buster as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./
RUN go build -mod=readonly -v -o server

# Application Container
FROM debian:buster-slim
ENV MONGO_URI=${{secrets.MONGO_URI}} ACCESS_SECRET=${{secrets.ACCESS_SECRET}}
RUN mkdir /app
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/server /app/server
CMD ["/app/server"]