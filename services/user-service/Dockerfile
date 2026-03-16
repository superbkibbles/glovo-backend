FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.work ./
COPY proto/ ./proto/
COPY pkg/ ./pkg/
COPY services/ ./services/

RUN cd pkg && go mod download
RUN cd services/user-service && go mod download

RUN cd services/user-service && CGO_ENABLED=0 GOOS=linux go build -o /user-service ./cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /user-service .

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 50052

CMD ["./user-service"]
