FROM arigaio/atlas:latest-community-distroless AS atlas
FROM mikefarah/yq:4 AS yq

FROM golang:1.27-alpine AS build
WORKDIR /src
COPY modules/platform ./modules/platform
COPY services/reservation ./services/reservation
WORKDIR /src/services/reservation
RUN go mod download
RUN CGO_ENABLED=0 go build -o /reservation ./cmd/reservation

FROM alpine:3.23
RUN adduser -D -u 10001 service
WORKDIR /app
COPY --from=build /reservation /usr/local/bin/reservation
COPY --from=atlas --chmod=755 /atlas /usr/local/bin/atlas
COPY --from=yq /usr/bin/yq /usr/local/bin/yq
COPY services/reservation/migrations /app/migrations
COPY --chmod=755 services/reservation/docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY services/reservation/config.docker.yaml /app/config.yaml
COPY services/reservation/migrations.yaml /app/migrations.yaml
USER service
EXPOSE 8070
ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["--config", "/app/config.yaml"]
