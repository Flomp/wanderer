from meilisearch import Client
import json

MEILISEARCH_URL='http://localhost:7700'
MEILISEARCH_API_KEY='p2gYZAWODOrwTPr4AYoahCZ9CI8y9bUd0yQLGk-E3m8'

client = Client(MEILISEARCH_URL, MEILISEARCH_API_KEY)

def init():
    client.create_index('cities500', {'primaryKey': 'id'})

    client.index('cities500').update_settings({
        'sortableAttributes': ['_geo',],
        'filterableAttributes': ['_geo']
        })

    client.create_index('trails', {'primaryKey': 'id'})

    client.index('trails').update_settings({
        'sortableAttributes': ['name', 'distance', 'elevation_gain', 'created',],
        'filterableAttributes': ['category', 'distance', 'elevation_gain', 'completed', '_geo', 'public', 'author']
        })


    json_file = open('cities500.json', encoding='utf-8')
    cities = json.load(json_file)

    client.index('cities500').add_documents(cities)

def generate_public_token():
    search_key = client.get_keys().results[0]

    search_rules = {
        'trails': {
            'filter': 'public=true'
        },
        'cities500': {}
    }
    token = client.generate_tenant_token(api_key_uid=search_key.uid, search_rules=search_rules, api_key=search_key.key)

    print(token)

generate_public_token()