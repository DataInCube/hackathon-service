FROM golang:1.24.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/hackathon-service ./cmd

FROM gcr.io/distroless/base-debian12

COPY --from=builder /bin/hackathon-service /hackathon-service
EXPOSE 8080

ENTRYPOINT ["/hackathon-service"]
