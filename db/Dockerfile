FROM golang:1.23-alpine as base
WORKDIR /

COPY migrations ./migrations

ARG TARGETARCH
COPY . .
RUN env GOOS=linux GOARCH=arm64 go build -o pocketbase
RUN chmod +x /pocketbase

FROM alpine:3.16 AS runtime
COPY --from=base /pocketbase ./pocketbase

ENV MEILI_URL=http://localhost:7700
ENV MEILI_MASTER_KEY=

EXPOSE 8090

ENTRYPOINT ["/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]