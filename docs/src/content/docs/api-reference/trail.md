---
title: trail
---

## POST file

POST /trail/{id}/file

> Body Parameters

```yaml
photos: string
gpx: string

```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» photos|body|string(binary)| no |none|
|» gpx|body|string(binary)| no |none|

> Response Examples

> 200 Response

```json
{}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|Success|Inline|

### Responses Data Schema

## GET show

GET /trail/{id}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "category": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "description": "string",
  "difficulty": "string",
  "distance": 0,
  "duration": 0,
  "elevation_gain": 0,
  "gpx": "string",
  "id": "string",
  "lat": 0,
  "location": "string",
  "lon": 0,
  "name": "string",
  "photos": [
    "string"
  ],
  "public": true,
  "summit_logs": [
    "string"
  ],
  "thumbnail": 0,
  "updated": "string",
  "waypoints": [
    "string"
  ]
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Record Not Found|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» author|string|true|none||none|
|» category|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» description|string|true|none||none|
|» difficulty|string|true|none||none|
|» distance|integer|true|none||none|
|» duration|integer|true|none||none|
|» elevation_gain|integer|true|none||none|
|» gpx|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» location|string|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» public|boolean|true|none||none|
|» summit_logs|[string]|true|none||none|
|» thumbnail|integer|true|none||none|
|» updated|string|true|none||none|
|» waypoints|[string]|true|none||none|

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» url|string|true|none||none|
|» status|integer|true|none||none|
|» response|object|true|none||none|
|»» code|integer|true|none||none|
|»» message|string|true|none||none|
|»» data|object|true|none||none|
|» isAbort|boolean|true|none||none|
|» originalError|object|true|none||none|
|»» url|string|true|none||none|
|»» status|integer|true|none||none|
|»» data|object|true|none||none|
|»»» code|integer|true|none||none|
|»»» message|string|true|none||none|
|»»» data|object|true|none||none|
|» name|string|true|none||none|

## POST update

POST /trail/{id}

> Body Parameters

```json
{
  "name": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» name|body|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "category": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "description": "string",
  "difficulty": "string",
  "distance": 0,
  "duration": 0,
  "elevation_gain": 0,
  "expand": {
    "category": {
      "collectionId": "string",
      "collectionName": "string",
      "created": "string",
      "id": "string",
      "img": "string",
      "name": "string",
      "updated": "string"
    },
    "waypoints": [
      {
        "author": "string",
        "collectionId": "string",
        "collectionName": "string",
        "created": "string",
        "description": "string",
        "icon": "string",
        "id": "string",
        "lat": 0,
        "lon": 0,
        "name": "string",
        "photos": [
          "string"
        ],
        "updated": "string"
      }
    ]
  },
  "gpx": "string",
  "id": "string",
  "lat": 0,
  "location": "string",
  "lon": 0,
  "name": "string",
  "photos": [
    "string"
  ],
  "public": true,
  "summit_logs": [
    "string"
  ],
  "thumbnail": 0,
  "updated": "string",
  "waypoints": [
    "string"
  ]
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Record Not Found|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» author|string|true|none||none|
|» category|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» description|string|true|none||none|
|» difficulty|string|true|none||none|
|» distance|integer|true|none||none|
|» duration|integer|true|none||none|
|» elevation_gain|integer|true|none||none|
|» expand|object|true|none||none|
|»» category|object|true|none||none|
|»»» collectionId|string|true|none||none|
|»»» collectionName|string|true|none||none|
|»»» created|string|true|none||none|
|»»» id|string|true|none||none|
|»»» img|string|true|none||none|
|»»» name|string|true|none||none|
|»»» updated|string|true|none||none|
|»» waypoints|[object]|true|none||none|
|»»» author|string|true|none||none|
|»»» collectionId|string|true|none||none|
|»»» collectionName|string|true|none||none|
|»»» created|string|true|none||none|
|»»» description|string|true|none||none|
|»»» icon|string|true|none||none|
|»»» id|string|true|none||none|
|»»» lat|number|true|none||none|
|»»» lon|number|true|none||none|
|»»» name|string|true|none||none|
|»»» photos|[string]|true|none||none|
|»»» updated|string|true|none||none|
|» gpx|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» location|string|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» public|boolean|true|none||none|
|» summit_logs|[string]|true|none||none|
|» thumbnail|integer|true|none||none|
|» updated|string|true|none||none|
|» waypoints|[string]|true|none||none|

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» url|string|true|none||none|
|» status|integer|true|none||none|
|» response|object|true|none||none|
|»» code|integer|true|none||none|
|»» message|string|true|none||none|
|»» data|object|true|none||none|
|» isAbort|boolean|true|none||none|
|» originalError|object|true|none||none|
|»» url|string|true|none||none|
|»» status|integer|true|none||none|
|»» data|object|true|none||none|
|»»» code|integer|true|none||none|
|»»» message|string|true|none||none|
|»»» data|object|true|none||none|
|» name|string|true|none||none|

## DELETE delete

DELETE /trail/{id}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "acknowledged": true
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|Record Not Found|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» acknowledged|boolean|true|none||none|

HTTP Status Code **404**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» url|string|true|none||none|
|» status|integer|true|none||none|
|» response|object|true|none||none|
|»» code|integer|true|none||none|
|»» message|string|true|none||none|
|»» data|object|true|none||none|
|» isAbort|boolean|true|none||none|
|» originalError|object|true|none||none|
|»» url|string|true|none||none|
|»» status|integer|true|none||none|
|»» data|object|true|none||none|
|»»» code|integer|true|none||none|
|»»» message|string|true|none||none|
|»»» data|object|true|none||none|
|» name|string|true|none||none|

## GET list

GET /trail

> Response Examples

> 200 Response

```json
{
  "page": 0,
  "perPage": 0,
  "totalItems": 0,
  "totalPages": 0,
  "items": [
    {
      "author": "string",
      "category": "string",
      "collectionId": "string",
      "collectionName": "string",
      "created": "string",
      "description": "string",
      "difficulty": "string",
      "distance": 0,
      "duration": 0,
      "elevation_gain": 0,
      "gpx": "string",
      "id": "string",
      "lat": 0,
      "location": "string",
      "lon": 0,
      "name": "string",
      "photos": [
        "string"
      ],
      "public": true,
      "summit_logs": [
        "string"
      ],
      "thumbnail": 0,
      "updated": "string",
      "waypoints": [
        "string"
      ]
    }
  ]
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» page|integer|true|none||none|
|» perPage|integer|true|none||none|
|» totalItems|integer|true|none||none|
|» totalPages|integer|true|none||none|
|» items|[object]|true|none||none|
|»» author|string|true|none||none|
|»» category|string|true|none||none|
|»» collectionId|string|true|none||none|
|»» collectionName|string|true|none||none|
|»» created|string|true|none||none|
|»» description|string|true|none||none|
|»» difficulty|string|true|none||none|
|»» distance|integer|true|none||none|
|»» duration|integer|true|none||none|
|»» elevation_gain|integer|true|none||none|
|»» gpx|string|true|none||none|
|»» id|string|true|none||none|
|»» lat|number|true|none||none|
|»» location|string|true|none||none|
|»» lon|number|true|none||none|
|»» name|string|true|none||none|
|»» photos|[string]|true|none||none|
|»» public|boolean|true|none||none|
|»» summit_logs|[string]|true|none||none|
|»» thumbnail|integer|true|none||none|
|»» updated|string|true|none||none|
|»» waypoints|[string]|true|none||none|

## PUT upload

PUT /trail/upload

> Body Parameters

```yaml
string

```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|Content-Type|header|string| yes |none|
|body|body|string(binary)| no |none|

> Response Examples

> 200 Response

```json
{
  "name": "string",
  "public": true,
  "difficulty": "string",
  "lat": 0,
  "lon": 0,
  "thumbnail": 0,
  "photos": [
    "string"
  ],
  "waypoints": [
    "string"
  ],
  "summit_logs": [
    "string"
  ],
  "expand": {
    "waypoints": [
      "string"
    ],
    "summit_logs": [
      "string"
    ]
  },
  "tags": [
    "string"
  ],
  "description": "string",
  "author": "string"
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|Bad Request|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» name|string|true|none||none|
|» public|boolean|true|none||none|
|» difficulty|string|true|none||none|
|» lat|number|true|none||none|
|» lon|number|true|none||none|
|» thumbnail|integer|true|none||none|
|» photos|[string]|true|none||none|
|» waypoints|[string]|true|none||none|
|» summit_logs|[string]|true|none||none|
|» expand|object|true|none||none|
|»» waypoints|[string]|true|none||none|
|»» summit_logs|[string]|true|none||none|
|» tags|[string]|true|none||none|
|» description|string|true|none||none|
|» author|string|true|none||none|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» url|string|true|none||none|
|» status|integer|true|none||none|
|» response|object|true|none||none|
|»» message|string|true|none||none|
|» isAbort|boolean|true|none||none|
|» originalError|object|true|none||none|
|»» status|integer|true|none||none|
|»» response|object|true|none||none|
|»»» message|string|true|none||none|
|» name|string|true|none||none|
