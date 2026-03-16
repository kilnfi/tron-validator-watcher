ARG NODE=node:24-alpine
ARG BUILDER=golang:1.26-alpine
ARG RUNNER=alpine:3.23

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

COPY . .
COPY --from=frontend-builder /workspace/internal/ui/dist ./internal/ui/dist

RUN apk --no-cache add gcc musl-dev

RUN go mod download \
  && go mod verify

RUN go build -v -o /usr/local/bin/tron-validator-watcher cmd/watcher/main.go

# Stage 3: Minimal runtime image
FROM ${RUNNER}

WORKDIR /home/tron

RUN apk --no-cache add ca-certificates curl \
  && update-ca-certificates

COPY --from=builder /usr/local/bin/tron-validator-watcher .

ENTRYPOINT ["./tron-validator-watcher"]
