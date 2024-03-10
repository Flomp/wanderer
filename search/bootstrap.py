from meilisearch import Client
from meilisearch.errors import MeilisearchApiError
import json
import os

MEILI_URL=os.getenv('MEILI_URL')
MEILI_MASTER_KEY=os.getenv('MEILI_MASTER_KEY')

client = Client(MEILI_URL, MEILI_MASTER_KEY)

def index_exists(index_name: str) -> bool:
    try:
        client.get_index(index_name)
        return True
    except MeilisearchApiError:
        return False

def init_indices():
    if not index_exists('cities500'):
        print("Creating cities index...")
        client.create_index('cities500', {'primaryKey': 'id'})

        client.index('cities500').update_settings({
            'sortableAttributes': ['_geo',],
            'filterableAttributes': ['_geo']
        })

        print("Starting data import...")
        json_file = open('cities500.json', encoding='utf-8')
        cities = json.load(json_file)

        client.index('cities500').add_documents(cities)
        print("Data import completed!")

    if not index_exists('trails'):
        print("Creating trails index...")
        client.create_index('trails', {'primaryKey': 'id'})

        client.index('trails').update_settings({
            'sortableAttributes': ['name', 'distance', 'elevation_gain', 'difficulty', 'created',],
            'filterableAttributes': ['category', 'difficulty', 'distance', 'elevation_gain', 'completed', '_geo', 'public', 'author']
        })


def generate_public_token():
    search_key = client.get_keys().results[0]

    search_rules = {
        'trails': {
            'filter': 'public=true'
        },
        'cities500': {}
    }
    token = client.generate_tenant_token(api_key_uid=search_key.uid, search_rules=search_rules, api_key=search_key.key)

    return token

print("Initializing indices...")
init_indices()
print("Indices initialized!")
token = generate_public_token()
print("Generating public token...")
print(f"PUBLIC_MEILISEARCH_API_TOKEN:{token}")

print("Bootstrapping complete!")
