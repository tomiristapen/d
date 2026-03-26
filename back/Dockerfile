FROM golang:1.23.4-alpine AS build

WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM alpine:3.21
WORKDIR /app
ENV PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng tesseract-ocr-data-rus
COPY --from=build /bin/api /app/api
COPY internal/platform/migrations /app/internal/platform/migrations
EXPOSE 8080
CMD ["/app/api"]
