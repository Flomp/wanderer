---
title: comment
---

## GET show

GET /comment/{id}

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
  "created": "string",
  "expand": {
    "author": {
      "avatar": "string",
      "collectionId": "string",
      "collectionName": "string",
      "created": "string",
      "email": "string",
      "emailVisibility": true,
      "id": "string",
      "token": "string",
      "updated": "string",
      "username": "string",
      "verified": true
    }
  },
  "id": "string",
  "rating": 0,
  "text": "string",
  "trail": "string",
  "updated": "string"
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
|» created|string|true|none||none|
|» expand|object|true|none||none|
|»» author|object|true|none||none|
|»»» avatar|string|true|none||none|
|»»» collectionId|string|true|none||none|
|»»» collectionName|string|true|none||none|
|»»» created|string|true|none||none|
|»»» email|string|true|none||none|
|»»» emailVisibility|boolean|true|none||none|
|»»» id|string|true|none||none|
|»»» token|string|true|none||none|
|»»» updated|string|true|none||none|
|»»» username|string|true|none||none|
|»»» verified|boolean|true|none||none|
|» id|string|true|none||none|
|» rating|integer|true|none||none|
|» text|string|true|none||none|
|» trail|string|true|none||none|
|» updated|string|true|none||none|

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

POST /comment/{id}

> Body Parameters

```json
{
  "text": "string",
  "trail": "string",
  "author": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» text|body|string| yes |none|
|» trail|body|string| yes |none|
|» author|body|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "id": "string",
  "rating": 0,
  "text": "string",
  "trail": "string",
  "updated": "string"
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
|» created|string|true|none||none|
|» id|string|true|none||none|
|» rating|integer|true|none||none|
|» text|string|true|none||none|
|» trail|string|true|none||none|
|» updated|string|true|none||none|

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

DELETE /comment/{id}

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

GET /comment

> Response Examples

> 200 Response

```json
[
  {
    "author": "string",
    "collectionId": "string",
    "collectionName": "string",
    "created": "string",
    "expand": {
      "author": {
        "avatar": "string",
        "collectionId": "string",
        "collectionName": "string",
        "created": "string",
        "email": "string",
        "emailVisibility": true,
        "id": "string",
        "token": "string",
        "updated": "string",
        "username": "string",
        "verified": true
      }
    },
    "id": "string",
    "rating": 0,
    "text": "string",
    "trail": "string",
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
|» expand|object|true|none||none|
|»» author|object|true|none||none|
|»»» avatar|string|true|none||none|
|»»» collectionId|string|true|none||none|
|»»» collectionName|string|true|none||none|
|»»» created|string|false|none||none|
|»»» email|string|false|none||none|
|»»» emailVisibility|boolean|false|none||none|
|»»» id|string|true|none||none|
|»»» token|string|false|none||none|
|»»» updated|string|false|none||none|
|»»» username|string|true|none||none|
|»»» verified|boolean|false|none||none|
|» id|string|true|none||none|
|» rating|integer|true|none||none|
|» text|string|true|none||none|
|» trail|string|true|none||none|
|» updated|string|true|none||none|

## PUT create

PUT /comment

> Body Parameters

```json
{
  "author": "string",
  "text": "string",
  "trail": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» author|body|string| yes |none|
|» text|body|string| yes |none|
|» trail|body|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "author": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "id": "string",
  "rating": 0,
  "text": "string",
  "trail": "string",
  "updated": "string"
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
|» created|string|true|none||none|
|» id|string|true|none||none|
|» rating|integer|true|none||none|
|» text|string|true|none||none|
|» trail|string|true|none||none|
|» updated|string|true|none||none|

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
