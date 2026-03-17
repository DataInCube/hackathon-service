FROM golang:1.24.3 AS builder

WORKDIR /app
COPY go.mod go.sum ./

RUN mkdir ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts && \
    update-ca-certificates --verbose --fresh

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GIT_TERMINAL_PROMPT=1

RUN git config --global url."git@github.com:".insteadOf "https://github.com/"
RUN go env -w GOPRIVATE=github.com/DataInCube/*
RUN --mount=type=ssh go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /bin/hackathon-service ./cmd

FROM gcr.io/distroless/base-debian12

COPY --from=builder /bin/hackathon-service /hackathon-service
EXPOSE 8080

ENTRYPOINT ["/hackathon-service"]
