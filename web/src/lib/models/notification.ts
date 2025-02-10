import type { UserAnonymous } from "./user";

enum NotificationType {
    trailCreate = "trail_create",
    trailShare = "trail_share",
    listCreate = "list_create",
    listShare = "list_share",
    newFollower = "new_follower",
    trailComment = "trail_comment"
};

interface Notification {
    id: string;
    type: NotificationType;
    metadata?: Record<string, any>;
    seen: boolean;
    recipient: string
    author: string
    created: string;
    expand: {
        recipient: UserAnonymous;
        author: UserAnonymous;
    }
}

export { type Notification, NotificationType }