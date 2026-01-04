FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -ldflags="-s -w"

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/quickserve /app/

EXPOSE 3000

ENTRYPOINT ["/app/quickserve", "--folder", "/data"]