---
title: list-share
---

## GET show

GET /list-share/{id}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "id": "string",
  "permission": "view",
  "list": "string",
  "updated": "2019-08-24T14:15:22Z",
  "user": "string"
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» id|string|true|none||none|
|» permission|string|true|none||none|
|» list|string|true|none||none|
|» updated|string(date-time)|true|none||none|
|» user|string|true|none||none|

#### Enum

|Name|Value|
|---|---|
|permission|view|
|permission|edit|

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

POST /list-share/{id}

> Body Parameters

```json
{
  "list": "string",
  "user": "string",
  "permission": "view"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» list|body|string| yes |none|
|» user|body|string| yes |none|
|» permission|body|string| yes |none|

#### Enum

|Name|Value|
|---|---|
|» permission|view|
|» permission|edit|

> Response Examples

> 200 Response

```json
{
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "id": "string",
  "permission": "view",
  "list": "string",
  "updated": "2019-08-24T14:15:22Z",
  "user": "string"
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» id|string|true|none||none|
|» permission|string|true|none||none|
|» list|string|true|none||none|
|» updated|string(date-time)|true|none||none|
|» user|string|true|none||none|

#### Enum

|Name|Value|
|---|---|
|permission|view|
|permission|edit|

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

GET /list-share

> Response Examples

> 200 Response

```json
[
  {
    "collectionId": "string",
    "collectionName": "string",
    "created": "string",
    "id": "string",
    "list": "string",
    "permission": "string",
    "updated": "string",
    "user": "string",
    "expand": {
      "user": {
        "avatar": "string",
        "collectionId": "string",
        "collectionName": "string",
        "id": "string",
        "username": "string"
      }
    }
  }
]
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» collectionId|string|false|none||none|
|» collectionName|string|false|none||none|
|» created|string|false|none||none|
|» id|string|false|none||none|
|» list|string|false|none||none|
|» permission|string|false|none||none|
|» updated|string|false|none||none|
|» user|string|false|none||none|
|» expand|object|false|none||none|
|»» user|object|true|none||none|
|»»» avatar|string|true|none||none|
|»»» collectionId|string|true|none||none|
|»»» collectionName|string|true|none||none|
|»»» id|string|true|none||none|
|»»» username|string|true|none||none|

## DELETE delete

DELETE /list-share/{id}

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
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|404|[Not Found](https://tools.ietf.org/html/rfc7231#section-6.5.4)|none|Inline|

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

## PUT create

PUT /list-share

> Body Parameters

```json
{
  "list": "string",
  "user": "string",
  "permission": "view"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» list|body|string| yes |none|
|» user|body|string| yes |none|
|» permission|body|string| yes |none|

#### Enum

|Name|Value|
|---|---|
|» permission|view|
|» permission|edit|

> Response Examples

> 200 Response

```json
{
  "collectionId": "string",
  "collectionName": "string",
  "created": "2019-08-24T14:15:22Z",
  "id": "string",
  "permission": "view",
  "list": "string",
  "updated": "2019-08-24T14:15:22Z",
  "user": "string"
}
```

### Responses

|HTTP Status Code |Meaning|Description|Data schema|
|---|---|---|---|
|200|[OK](https://tools.ietf.org/html/rfc7231#section-6.3.1)|none|Inline|
|400|[Bad Request](https://tools.ietf.org/html/rfc7231#section-6.5.1)|none|Inline|

### Responses Data Schema

HTTP Status Code **200**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string(date-time)|true|none||none|
|» id|string|true|none||none|
|» permission|string|true|none||none|
|» list|string|true|none||none|
|» updated|string(date-time)|true|none||none|
|» user|string|true|none||none|

#### Enum

|Name|Value|
|---|---|
|permission|view|
|permission|edit|

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

# Data Schema

