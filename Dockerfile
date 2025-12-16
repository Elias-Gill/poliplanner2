ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /run-app .

FROM debian:bookworm

# Base dir
ENV APP_BASE_DIR=/var/poliplanner

# Copy binary
COPY --from=builder /run-app /usr/local/bin/run-app

# Copy Assets / files
COPY --from=builder /usr/src/app/internal/db/migrations \
    /var/poliplanner/internal/db/migrations

COPY --from=builder /usr/src/app/internal/excelparser/layouts \
    /var/poliplanner/internal/excelparser/layouts

COPY --from=builder /usr/src/app/internal/excelparser/metadata \
    /var/poliplanner/internal/excelparser/metadata

COPY --from=builder /usr/src/app/web \
            /var/poliplanner/web

WORKDIR /var/poliplanner

CMD ["run-app"]
