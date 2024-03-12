FROM golang:1.22-alpine
RUN apk add build-base

WORKDIR /

COPY go.mod ./
COPY go.sum ./
COPY migrations ./migrations
RUN tar -xvf migrations/initial_data/cities500.tar.gz -C migrations/initial_data

RUN go mod download

COPY *.go ./

RUN go build -o /pocketbase

ENV MEILI_URL=http://localhost:7700
ENV MEILI_MASTER_KEY=

EXPOSE 8090

ENTRYPOINT ["/pocketbase", "serve", "--http=0.0.0.0:8090", "--dir=/pb_data"]