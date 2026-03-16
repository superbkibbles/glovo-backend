FROM golang:1.23-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY pkg/ ./pkg/
COPY proto/ ./proto/
COPY cmd/seed/ ./cmd/seed/

RUN cd pkg && go mod download
RUN cd cmd/seed && go mod download

RUN cd cmd/seed && CGO_ENABLED=0 GOOS=linux go build -o /seed .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /seed .

CMD ["./seed"]
