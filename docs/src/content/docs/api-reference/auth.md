---
title: auth
---

# auth

## POST login

POST /auth/login

> Body Parameters

```json
{
  "username": "string",
  "password": "string"
}
```

### Params

|Name|Location|Type|Required|Description|
|---|---|---|---|---|
|Content-Type|header|string| yes |none|
|body|body|object| no |none|
|» username|body|string| yes |none|
|» password|body|string| yes |none|

> Response Examples

> 200 Response

```json
{
  "record": {
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
  },
  "token": "string"
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
|» record|object|true|none||none|
|»» avatar|string|true|none||none|
|»» collectionId|string|true|none||none|
|»» collectionName|string|true|none||none|
|»» created|string|true|none||none|
|»» email|string|true|none||none|
|»» emailVisibility|boolean|true|none||none|
|»» id|string|true|none||none|
|»» language|string|true|none||none|
|»» location|object|true|none||none|
|»»» lat|number|true|none||none|
|»»» lon|number|true|none||none|
|»»» name|string|true|none||none|
|»» token|string|true|none||none|
|»» unit|string|true|none||none|
|»» updated|string|true|none||none|
|»» username|string|true|none||none|
|»» verified|boolean|true|none||none|
|» token|string|true|none||none|

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
