trap "kill 0" EXIT

export MEILI_URL=https://meilisearch.lzsh.work
export MEILI_MASTER_KEY=fk.PYjRArYTHie1XHqZ3t0~kBIabwcEY34YH74yROgXDi5_Q~TeQOsngUwevTDH2JcEp4qP3
export PUBLIC_POCKETBASE_URL=http://10.10.10.35:22380
export PUBLIC_DISABLE_SIGNUP=false
export UPLOAD_USER=
export UPLOAD_PASSWORD=

# cd search && meilisearch --master-key $MEILI_MASTER_KEY &
# cd db && ./pocketbase serve &
# cd web && node build

# wait
npm run dev
