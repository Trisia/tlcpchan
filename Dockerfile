FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git nodejs npm

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY tlcpchan/ ./tlcpchan/
COPY tlcpchan-cli/ ./tlcpchan-cli/
COPY tlcpchan-ui/ ./tlcpchan-ui/

WORKDIR /app/tlcpchan
RUN go build -ldflags="-s -w" -o /app/bin/tlcpchan ./cmd/tlcpchan

WORKDIR /app/tlcpchan-cli
RUN go build -ldflags="-s -w" -o /app/bin/tlcpchan-cli .

WORKDIR /app/tlcpchan-ui/web
RUN npm install && npm run build
WORKDIR /app/tlcpchan-ui
RUN go build -ldflags="-s -w" -o /app/bin/tlcpchan-ui .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/tlcpchan /app/bin/tlcpchan-cli /app/bin/tlcpchan-ui /usr/local/bin/
COPY --from=builder /app/tlcpchan-ui/dist /app/ui/dist

RUN mkdir -p /app/config /app/certs/tlcp /app/certs/tls /app/logs

EXPOSE 30080 30000

ENV TZ=Asia/Shanghai

CMD ["tlcpchan"]
