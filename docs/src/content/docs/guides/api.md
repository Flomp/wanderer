---
title: wanderer API
description: How to use wanderer's API
---

wanderer comes with a fully functional RESTful API out of the box. It largely follows the CRUD conventions implemented by the [PocketBase backend](https://pocketbase.io/docs/api-records/#crud-actions). All endpoints are available at `http://<wanderer_host>/api/v1`. The full technical API documentation can be found at `http://<wanderer_host>/docs/api/index.html` and is also provided in the [API Reference](/api-reference/auth).

## Authentication
wanderer's API uses cookie-based authentication. To receive an auth cookie send a request to the `/auth/login` endpoint. The request must contain a JSON body in the following form:
```
{
    username: string,
    password: string
}
```
The cookie from the response can be sent in a subsequent request to authenticate it.
### Example
```bash
curl --header "Content-Type: application/json" --request POST \
--cookie-jar ./wanderer-credentials \
--data '{"username":"MyUser","password":"mysecretpassword"}' \
http://localhost:3000/api/v1/auth/login
```


## Upload trails
One common use case for wanderer's API is bulk uploading GPX files to create new trails. For that, the API provides a separate endpoint: `/trail/upload`. You must first log in to use the endpoint. Afterwards you can send a GPX file to the endpoint to let wanderer parse it an create a new trail in your collection. wanderer will try to infer as much information as possible from the file itself. All additional information can be added to the trail via the UPDATE [endpoint](/api-reference/trail#post-update).

### Example
```bash
curl --location --request PUT 'http://localhost:3000/api/v1/trail/upload' \
--header 'Content-Type: application/gpx+xml' \
--cookie './wanderer-credentials' \
--data-binary '@my_trail.gpx'
```
