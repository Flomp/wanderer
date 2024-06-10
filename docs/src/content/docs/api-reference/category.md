---
title: category
---

## GET list

GET /category

> Response Examples

> 200 Response

```json
[
  {
    "collectionId": "string",
    "collectionName": "string",
    "created": "string",
    "id": "string",
    "img": "string",
    "name": "string",
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
|» collectionId|string|true|none||none|
|» collectionName|string|true|none||none|
|» created|string|true|none||none|
|» id|string|true|none||none|
|» img|string|true|none||none|
|» name|string|true|none||none|
|» updated|string|true|none||none|
