FROM alpine:3.16

WORKDIR /

COPY migrations ./migrations

COPY ./pocketbase_arm64 /pocketbase
RUN chmod +x /pocketbase

ENV MEILI_URL=http://localhost:7700
ENV MEILI_MASTER_KEY=

EXPOSE 8090

ENTRYPOINT ["/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]