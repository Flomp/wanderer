---
title: Custom categories
description: How to create custom trail categories
---

wanderer uses categories to classify what kind of activity a trail belongs to. Out of the box you get: Biking, Canoeing, Climbing, Hiking, Skiing and Walking. However, you can adapt these categories to your needs or add completely new ones.

## Backend access

First, you need access to the PocketBase backend. If you are using docker make sure to forward the internal port 8090 to a public port. With the default configuration, the PocketBase admin panel is available at `http://localhost:8090/_/`. If this is your first time visiting the panel you will need to create an admin account. 
To create backend access navigate to your docker-compose.yaml file and type:
```
docker compose exec -it db /pocketbase superuser upsert email@example.com myverysecurepassword
```
Now, you will have access with the user "email@example.com" and the password "myverysecurepassword" to all tables in the backend and can modify the underlying data directly.

## Modifying categories

![Pocketbase Categories](../../../assets/guides/pocketbase_categories.png)

In the PocketBase admin panel, click on the `categories` table in the list on the left side. All existing categories will be listed here. To edit one simply click on the row, edit the data you want to change, and click "Save". To delete a category check the box at the beginning of the row and click "Delete selected". To create a new category click the "New record" button in the top right corner, give your new category a name and a background image, and click "Save".
