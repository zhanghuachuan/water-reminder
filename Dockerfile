FROM golang:1.21 as builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /water-reminder

FROM alpine:latest
WORKDIR /app
COPY --from=builder /water-reminder .
COPY config/config.yaml .

EXPOSE 8000
CMD ["./water-reminder"]