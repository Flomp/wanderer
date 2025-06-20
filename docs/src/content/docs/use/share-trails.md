---
title: Share trails
description: How to share trails with other users
---

wanderer allows you to share your trails with other users. You can either publish you trail making it accessible for everyone or share it with specific users. To get started head over to `/trails` and select the trail you want to share or publish.

## Publish a trail

From the <span class="inline-block w-8 h-8 bg-primary rounded-full text-center text-white">⋮</span> menu select "Edit". In the panel on the right toggle the "Public" switch to on and save the trail. Your trail is now public and everyone can see it. Even people without an account.


## Share a trail

![Share trail](../../../assets/guides/wanderer_share.gif)

If you want to be more particular about who can see your trail you can instead share your trail. From the <span class="inline-block w-8 h-8 bg-primary rounded-full text-center text-white">⋮</span> menu select "Share". In the dialog, search for the user you want to share your trail with. You can now choose the permission the user should have. You can choose between "View" or "Edit". A user with "Edit" permission can change all data (including the route) of the trail.

If you no longer want to share the trail with a user, simply click the red trashcan icon next to their name.

:::note
Wanderer supports trail sharing between users on different instances (servers), thanks to its federated design. However, there are important limitations to be aware of:

- **The trail must be public** in order to be shareable with users on other instances.
- **Shared trails are view-only**: The user you share it with will be able to view the trail and engage with it (like or comment), but **they cannot edit it**.
- Sharing a trail with another user is similar to a **mention** in the fediverse—it notifies them and gives them visibility, but does not grant collaborative access.

If you're looking for true collaboration on a trail (such as shared editing), both users must be on the same instance.
:::