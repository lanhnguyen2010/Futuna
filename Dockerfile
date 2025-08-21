# Build stage
FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o analyzer cmd/analyzer/main.go
RUN CGO_ENABLED=0 go build -o web cmd/web/main.go

# Final image
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=builder /app/analyzer /app/analyzer
COPY --from=builder /app/web /app/web
COPY web/static web/static
CMD ["/app/web"]
