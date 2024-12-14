FROM golang:1.23.4-alpine3.21 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build main.go .

FROM scratch
COPY --from=builder /app/main /app/main
ENTRYPOINT ["/app/main"]