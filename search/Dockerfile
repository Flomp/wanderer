FROM getmeili/meilisearch:v1.6.2

COPY ./migrations/migration.dump /meili_data/dumps/migration.dump
COPY ./entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

CMD [ "/entrypoint.sh" ]


