# syntax=docker/dockerfile:1
FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.work ./
COPY proto/ ./proto/
COPY pkg/ ./pkg/
COPY services/ ./services/
RUN --mount=type=cache,target=/go/pkg/mod \
    cd pkg && go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    cd services/settings-service && go mod download
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd services/settings-service && CGO_ENABLED=0 GOOS=linux go build -o /settings-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /settings-service .
RUN adduser -D -g '' appuser
USER appuser
EXPOSE 50058
CMD ["./settings-service"]
