FROM node:slim AS runtime
WORKDIR /app

COPY ./dist dist/
COPY package*.json .

RUN npm ci --omit=dev

ENV HOST=0.0.0.0
ENV PORT=4321
EXPOSE 4321
CMD node ./dist/server/entry.mjs