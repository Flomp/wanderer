---
title: user
---

## POST file

POST /user/{id}/file

> Body Parameters

```yaml
avatar: string

```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» avatar|body|string(binary)| no |none|

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

GET /user/{id}

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "avatar": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "email": "string",
  "emailVisibility": true,
  "id": "string",
  "language": "string",
  "location": {
    "lat": 0,
    "lon": 0,
    "name": "string"
  },
  "token": "string",
  "unit": "string",
  "updated": "string",
  "username": "string",
  "verified": true
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
|» avatar|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» email|string|true|none||none|
|» emailVisibility|boolean|true|none||none|
|» id|string|true|none||none|
|» language|string|true|none||none|
|» location|object|true|none||none|
|»» lat|number|true|none||none|
|»» lon|number|true|none||none|
|»» name|string|true|none||none|
|» token|string|true|none||none|
|» unit|string|true|none||none|
|» updated|string|true|none||none|
|» username|string|true|none||none|
|» verified|boolean|true|none||none|

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

POST /user/{id}

> Body Parameters

```json
{}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|id|path|string| yes |none|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|

> Response Examples

> 200 Response

```json
{
  "avatar": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "email": "string",
  "emailVisibility": true,
  "id": "string",
  "language": "string",
  "location": {
    "lat": 0,
    "lon": 0,
    "name": "string"
  },
  "token": "string",
  "unit": "string",
  "updated": "string",
  "username": "string",
  "verified": true
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
|» avatar|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» email|string|true|none||none|
|» emailVisibility|boolean|true|none||none|
|» id|string|true|none||none|
|» language|string|true|none||none|
|» location|object|true|none||none|
|»» lat|number|true|none||none|
|»» lon|number|true|none||none|
|»» name|string|true|none||none|
|» token|string|true|none||none|
|» unit|string|true|none||none|
|» updated|string|true|none||none|
|» username|string|true|none||none|
|» verified|boolean|true|none||none|

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

DELETE /user/{id}

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

## PUT create

PUT /user

> Body Parameters

```json
{
  "username": "string",
  "password": "string",
  "passwordConfirm": "string",
  "email": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|body|body|object| no |none|
|» username|body|string| yes |none|
|» password|body|string| yes |none|
|» passwordConfirm|body|string| yes |none|
|» email|body|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "avatar": "string",
  "collectionId": "string",
  "collectionName": "string",
  "created": "string",
  "emailVisibility": true,
  "id": "string",
  "language": "string",
  "location": null,
  "token": "string",
  "unit": "string",
  "updated": "string",
  "username": "string",
  "verified": true
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
|» avatar|string|true|none||none|
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» emailVisibility|boolean|true|none||none|
|» id|string|true|none||none|
|» language|string|true|none||none|
|» location|null|true|none||none|
|» token|string|true|none||none|
|» unit|string|true|none||none|
|» updated|string|true|none||none|
|» username|string|true|none||none|
|» verified|boolean|true|none||none|

HTTP Status Code **400**

|Name|Type|Required|Restrictions|Title|description|
|---|---|---|---|---|---|
|» url|string|true|none||none|
|» status|integer|true|none||none|
|» response|object|true|none||none|
|»» code|integer|true|none||none|
|»» message|string|true|none||none|
|»» data|object|true|none||none|
|»»» passwordConfirm|object|true|none||none|
|»»»» code|string|true|none||none|
|»»»» message|string|true|none||none|
|»»» username|object|true|none||none|
|»»»» code|string|true|none||none|
|»»»» message|string|true|none||none|
|» isAbort|boolean|true|none||none|
|» originalError|object|true|none||none|
|»» url|string|true|none||none|
|»» status|integer|true|none||none|
|»» data|object|true|none||none|
|»»» code|integer|true|none||none|
|»»» message|string|true|none||none|
|»»» data|object|true|none||none|
|»»»» passwordConfirm|object|true|none||none|
|»»»»» code|string|true|none||none|
|»»»»» message|string|true|none||none|
|»»»» username|object|true|none||none|
|»»»»» code|string|true|none||none|
|»»»»» message|string|true|none||none|
|» name|string|true|none||none|
