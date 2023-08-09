FROM golang:1.21-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY web/ ./web/
COPY scripts/ ./scripts/

RUN apk add --no-cache build-base

RUN CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o /rsslay cmd/rsslay/main.go

FROM alpine:latest

LABEL org.opencontainers.image.title="rsslay"
LABEL org.opencontainers.image.source=https://github.com/piraces/rsslay
LABEL org.opencontainers.image.description="rsslay turns RSS or Atom feeds into Nostr profiles"
LABEL org.opencontainers.image.authors="Raúl Piracés"
LABEL org.opencontainers.image.licenses=MIT

ENV PORT="8080"
ENV DB_DIR="/db/rsslay.sqlite"
ENV DEFAULT_PROFILE_PICTURE_URL="https://i.imgur.com/MaceU96.png"
ENV SECRET="CHANGE_ME"
ENV VERSION=0.5.4
ENV REPLAY_TO_RELAYS=false
ENV RELAYS_TO_PUBLISH_TO=""
ENV NITTER_INSTANCES=""
ENV DEFAULT_WAIT_TIME_BETWEEN_BATCHES=60000
ENV DEFAULT_WAIT_TIME_FOR_RELAY_RESPONSE=3000
ENV MAX_EVENTS_TO_REPLAY=10
ENV ENABLE_AUTO_NIP05_REGISTRATION=false
ENV MAIN_DOMAIN_NAME=""
ENV OWNER_PUBLIC_KEY=""
ENV MAX_SUBROUTINES=20
ENV INFO_RELAY_NAME="rsslay"
ENV INFO_CONTACT=""
ENV MAX_CONTENT_LENGTH=250
ENV LOG_LEVEL="WARN"
ENV DELETE_FAILING_FEEDS=false
ENV REDIS_CONNECTION_STRING=""

COPY --from=build /rsslay .
COPY --from=build /app/web/assets/ ./web/assets/

CMD [ "/rsslay" ]