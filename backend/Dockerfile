# Backend MedClaim — build 3 binary (server, migrate, seed) lalu jalankan
# migrate -> seed -> server saat container start.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/migrate ./cmd/migrate \
 && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/seed ./cmd/seed

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/ ./
EXPOSE 8081
CMD ["/bin/sh", "-c", "./migrate && ./seed && ./server"]
