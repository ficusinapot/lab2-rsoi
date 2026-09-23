FROM golang:1.27-alpine AS build
WORKDIR /src
COPY modules/platform ./modules/platform
COPY services/gateway ./services/gateway
WORKDIR /src/services/gateway
RUN go mod download
RUN CGO_ENABLED=0 go build -o /gateway ./cmd/gateway
FROM alpine:3.23
RUN adduser -D -u 10001 service
COPY --from=build /gateway /usr/local/bin/gateway
COPY services/gateway/config.docker.yaml /app/config.yaml
USER service
EXPOSE 8080
ENTRYPOINT ["gateway"]
CMD ["--config", "/app/config.yaml"]
