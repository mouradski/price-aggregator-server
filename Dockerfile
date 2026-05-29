# Build stage
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache dependencies first
COPY go.mod go.sum ./
RUN go mod download

# Build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/aggregator ./cmd/aggregator

# Runtime stage: minimal, includes CA certificates for outbound TLS (wss/https)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/aggregator /aggregator

EXPOSE 8090

ENTRYPOINT ["/aggregator"]
