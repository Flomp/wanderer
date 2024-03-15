#!/bin/sh

if [ -z "$(ls -A /meili_data/data.ms/indexes)" ]; then
   meilisearch --import-dump /meili_data/dumps/migration.dump
else
   meilisearch
fi
