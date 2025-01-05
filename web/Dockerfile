FROM node:22-alpine AS base
WORKDIR /app

COPY package.json package-lock.json ./

FROM base AS prod-deps
RUN npm install --omit=dev

FROM base AS build-deps
RUN npm install

FROM build-deps AS build
ENV PUBLIC_VALHALLA_URL=https://valhalla1.openstreetmap.de
COPY . .
RUN npm run build

FROM base AS runtime
COPY --from=prod-deps /app/node_modules ./node_modules
COPY --from=build /app/build ./build

EXPOSE 3000
ENV NODE_ENV=production

COPY ./cron.sh /etc/periodic/15min/cron
RUN chmod +x /etc/periodic/15min/cron
RUN mkdir /app/uploads
RUN apk add curl

ENV PUBLIC_POCKETBASE_URL=http://127.0.0.1:8090
ENV MEILI_URL=http://127.0.0.1:7700
ENV PUBLIC_DISABLE_SIGNUP=false
ENV UPLOAD_USER=
ENV UPLOAD_PASSWORD=

CMD crond && node build