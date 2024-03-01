FROM golang:1.22-alpine
RUN apk add build-base

WORKDIR /

COPY go.mod ./
COPY go.sum ./
COPY pb_migrations ./pb_migrations

RUN go mod download

COPY *.go ./

RUN go build -o /pocketbase
RUN /pocketbase migrate

ENV PUBLIC_MEILISEARCH_URL=http://localhost:7700
ENV MEILISEARCH_MASTER_KEY=

EXPOSE 8090

ENTRYPOINT ["/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]