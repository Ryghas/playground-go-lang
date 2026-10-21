# ---- Builder ----
FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/pokemon-service ./cmd

# ---- Runtime ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/pokemon-service /pokemon-service
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/pokemon-service"]
