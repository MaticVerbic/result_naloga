FROM golang:1.15 AS build-env
WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /tmp/api ./cmd/api/

FROM alpine:latest
WORKDIR /app
RUN apk update && apk upgrade && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=build-env /tmp/api /app/api
CMD ["/app/api"]
EXPOSE 80
HEALTHCHECK --interval=10s --timeout=5s \
  CMD if [ $(curl -s -o /dev/null -w "%{http_code}" -f -I -m 4 -A "HEALTHCHECK" http://localhost/ping) = 200 ]; then exit 0; else exit 1; fi
