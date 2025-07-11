import type { Actor } from "./activitypub/actor";
import type { List } from "./list";
import type { SummitLog } from "./summit_log";
import type { Trail } from "./trail";

interface FeedItem {
    id: string;
    actor: string;
    author: string;
    item: string;
    type: "trail" | "list" | "summit_log"
    expand: {
        actor?: Actor,
        author?: Actor,
        item: Trail | List
    }
    created: string;

}

export { type FeedItem }