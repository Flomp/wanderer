import type { Actor } from "./activitypub/actor";

enum NotificationType {
    trailShare = "trail_share",
    trailLike = "trail_like",
    listShare = "list_share",
    newFollower = "new_follower",
    trailComment = "trail_comment",
    summitLogCreate = "summit_log_create",
    commentMention = "comment_mention",
    trailMention = "trail_mention",
    summitLogMention = "summit_log_mention",

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
        recipient: Actor;
        author: Actor;
    }
}

export { NotificationType, type Notification };
