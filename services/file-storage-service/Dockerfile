FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work ./
COPY proto/ ./proto/
COPY pkg/ ./pkg/
COPY services/ ./services/

RUN cd pkg && go mod download
RUN cd services/file-storage-service && go mod download

RUN cd services/file-storage-service && CGO_ENABLED=0 GOOS=linux go build -o /file-storage-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /file-storage-service .

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 50054

CMD ["./file-storage-service"]
