# syntax=docker/dockerfile:1

FROM golang:1.19-alpine as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY assets/* ./assets/
COPY templates/* ./templates/

RUN apk add --no-cache build-base

RUN CGO_ENABLED=1 go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o /rsslay

FROM alpine:latest

LABEL org.opencontainers.image.title="rsslay"
LABEL org.opencontainers.image.source=https://github.com/piraces/rsslay
LABEL org.opencontainers.image.description="rsslay turns RSS or Atom feeds into Nostr profiles"
LABEL org.opencontainers.image.authors="Raúl Piracés"
LABEL org.opencontainers.image.licenses=MIT

ENV PORT="8080"
ENV DB_DIR="/db"
ENV DEFAULT_PROFILE_PICTURE_URL="https://i.imgur.com/MaceU96.png"
ENV SECRET="CHANGE_ME"
ENV VERSION=0.2.1

COPY --from=build /rsslay .

CMD [ "/rsslay" ]