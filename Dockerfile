#  Builder
ARG BUILDER=golang:1.24.1-alpine
ARG RUNNER=alpine:3.21

FROM ${BUILDER} AS builder

WORKDIR /workspace

COPY . .

RUN apk --no-cache add gcc musl-dev

RUN go mod download \
  && go mod verify

RUN go build -v -o /usr/local/bin/tron-validator-watcher cmd/watcher/main.go

FROM ${RUNNER}

WORKDIR /home/tron

RUN apk --no-cache add ca-certificates curl \
  && update-ca-certificates

COPY --from=builder /usr/local/bin/tron-validator-watcher .

ENTRYPOINT ["./tron-validator-watcher"]