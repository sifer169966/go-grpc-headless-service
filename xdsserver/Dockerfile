# Defining App builder image
FROM golang:1.21-alpine3.17 AS builder

RUN apk update; \
    apk add --no-cache

# Define current working directory
WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . ./

# Build App
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o .bin/xdsserver ./xdsserver/main.go

RUN ls -a
# Defining App image
FROM alpine:3.17 as release

RUN apk add --no-cache --update ca-certificates

# Copy App binary to image
COPY --from=builder /app/.bin/xdsserver /app/cmd/
RUN chmod +x /app/cmd/xdsserver

ARG NONROOT_GROUP=nonroot-group
ARG NONROOT_USER=nonroot-user
ARG USER_ID=20000

RUN addgroup -S $NONROOT_GROUP && adduser -S -u $USER_ID $NONROOT_USER -G $NONROOT_GROUP

USER $NONROOT_USER:$NONROOT_GROUP

WORKDIR /app
RUN ls -a
CMD ["cmd/xdsserver"]
