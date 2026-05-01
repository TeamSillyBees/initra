FROM golang:1.26 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/initra ./cmd/server

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=builder /out/initra /app/initra
COPY --from=builder /app/configs /app/configs
COPY --from=builder /app/db /app/db

EXPOSE 8080

ENTRYPOINT ["/app/initra"]
