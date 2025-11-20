# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the service specified by SERVICE arg
ARG SERVICE
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/${SERVICE}

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/app .

EXPOSE 8080
EXPOSE 8081

CMD ["./app"]
