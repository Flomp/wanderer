---
title: waypoint
---

## GET show

GET /waypoint/{id}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "description": "string",
  "icon": "string",
  "id": "string",
  "lat": 0,
  "lon": 0,
  "name": "string",
  "photos": [
    "string"
  ],
  "updated": "2019-08-24T14:15:22Z"
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
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» description|string|true|none||none|
|» icon|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» updated|string(date-time)|true|none||none|

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

POST /waypoint/{id}

> Body Parameters

```json
{
  "name": "string",
  "description": "string",
  "lat": 0,
  "lon": 0,
  "icon": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» name|body|string| no |none|
|» description|body|string| no |none|
|» lat|body|number| yes |none|
|» lon|body|number| yes |none|
|» icon|body|string| no |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "description": "string",
  "icon": "string",
  "id": "string",
  "lat": 0,
  "lon": 0,
  "name": "string",
  "photos": [
    "string"
  ],
  "updated": "2019-08-24T14:15:22Z"
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
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» description|string|true|none||none|
|» icon|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» updated|string(date-time)|true|none||none|

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

DELETE /waypoint/{id}

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

GET /waypoint

> Response Examples

> 200 Response

```json
[
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
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|OK|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» author|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» description|string|true|none||none|
|» icon|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» updated|string|true|none||none|

## PUT create

PUT /waypoint

> Body Parameters

```json
{
  "name": "string",
  "description": "string",
  "author": "string",
  "lat": 0,
  "lon": 0,
  "icon": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» name|body|string| no |none|
|» description|body|string| no |none|
|» author|body|string| yes |none|
|» lat|body|number| yes |none|
|» lon|body|number| yes |none|
|» icon|body|string| no |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "description": "string",
  "icon": "string",
  "id": "string",
  "lat": 0,
  "lon": 0,
  "name": "string",
  "photos": [
    "string"
  ],
  "updated": "2019-08-24T14:15:22Z"
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
|» author|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» description|string|true|none||none|
|» icon|string|true|none||none|
|» id|string|true|none||none|
|» lat|number|true|none||none|
|» lon|number|true|none||none|
|» name|string|true|none||none|
|» photos|[string]|true|none||none|
|» updated|string(date-time)|true|none||none|

HTTP Status Code **400**

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
