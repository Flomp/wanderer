FROM node:18-alpine
WORKDIR /app
COPY ./build build/
COPY package*.json .
RUN npm ci --omit=dev
EXPOSE 3000
ENV NODE_ENV=production

ENV PUBLIC_POCKETBASE_URL=http://127.0.0.1:8090
ENV PUBLIC_MEILISEARCH_URL=http://127.0.0.1:7700
ENV PUBLIC_MEILISEARCH_API_TOKEN=

CMD [ "node", "build" ]