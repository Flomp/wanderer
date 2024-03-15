FROM getmeili/meilisearch:v1.6.2

COPY ./migrations/migration.dump /meili_data/dumps/migration.dump
COPY ./run.sh /run.sh

RUN chmod +x /run.sh

CMD [ "/run.sh" ]


