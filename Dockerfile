ARG NODE=node:24-alpine@sha256:116b800c3754b440ee0aa3ed2f20f5bf3d99a554bfd9adeea09ea8017f6e85f5
ARG BUILDER=golang:1.26-alpine@sha256:ecb67f987aa1a57794c91d19ade7ae0209f2f0f5496aaef9e6ae85ddf92d02ab
ARG RUNNER=alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

# Stage 1: Build frontend
FROM ${NODE} AS frontend-builder

WORKDIR /workspace/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary (embeds the frontend dist)
FROM ${BUILDER} AS builder

WORKDIR /workspace

RUN apk --no-cache add gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
COPY --from=frontend-builder /workspace/internal/ui/dist ./internal/ui/dist

RUN go build -v -o /usr/local/bin/tron-validator-watcher cmd/watcher/main.go

# Stage 3: Minimal runtime image
FROM ${RUNNER}

WORKDIR /home/tron

RUN apk --no-cache upgrade \
  && apk --no-cache add ca-certificates \
  && update-ca-certificates

COPY --from=builder /usr/local/bin/tron-validator-watcher .

ENTRYPOINT ["./tron-validator-watcher"]
