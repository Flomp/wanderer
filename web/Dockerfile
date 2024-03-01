FROM node:18-alpine AS build
WORKDIR /app
COPY package*.json .
RUN npm ci
COPY . .
RUN npm run build
RUN npm prune --production

FROM node:18-alpine
WORKDIR /app
COPY --from=build /app/build build/
COPY --from=build /app/node_modules node_modules/
COPY package.json .
EXPOSE 3000
ENV NODE_ENV=production

ENV PUBLIC_POCKETBASE_URL=http://127.0.0.1:8090
ENV PUBLIC_MEILISEARCH_URL=http://127.0.0.1:7700
ENV PUBLIC_MEILISEARCH_API_TOKEN=

CMD [ "node", "build" ]