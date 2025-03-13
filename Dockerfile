FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o quickserve main.go

FROM gcr.io/distroless/static-debian12

WORKDIR /app

COPY --from=builder /app/quickserve /app/

RUN mkdir /data

EXPOSE 3000

ENTRYPOINT ["/app/quickserve", "--folder", "/data"]