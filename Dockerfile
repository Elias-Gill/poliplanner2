ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

# Install Node.js y npm
RUN apt-get update \
    && apt-get install -y --no-install-recommends curl ca-certificates \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /run-app . \
    && npm install \
    && npm run build:css

FROM debian:bookworm

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y sqlite3 vim

ENV APP_BASE_DIR=/var/poliplanner

COPY --from=builder /usr/src/app/internal/ /var/poliplanner/internal

COPY --from=builder /usr/src/app/web /var/poliplanner/web

COPY --from=builder /run-app /usr/local/bin/run-app

WORKDIR /var/poliplanner

CMD ["run-app"]
