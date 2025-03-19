import { error, type NumericRange, type RequestEvent } from "@sveltejs/kit";
import { ClientResponseError, type ListResult } from "pocketbase";
import { pb } from "$lib/pocketbase";
import { ZodError, type ZodSchema } from "zod";
import { RecordListOptionsSchema, RecordIdSchema, RecordOptionsSchema } from "$lib/models/api/base_schema";

export class APIError extends Error {
    status: number;
    message: string;
    detail: any;

    constructor(status: number, message: string, detail?: any) {
        super();
        this.status = status;
        this.message = message;
        this.detail = detail
    }
}


export enum Collection {
    users = "users",
    categories = "categories",
    comments = "comments",
    follows = "follows",
    integrations = "integrations",
    list_share = "list_share",
    lists = "lists",
    notifications = "notifications",
    settings = "settings",
    summit_logs = "summit_logs",
    trail_share = "trail_share",
    trails = "trails",
    tags = "tags",
    waypoints = "waypoints",
    activities = "activities",
    follow_counts = "follow_counts",
    trails_bounding_box = "trails_bounding_box",
    trails_filter = "trails_filter",
    users_anonymous = "users_anonymous"
}

export async function list<T>(event: RequestEvent, collection: Collection) {
    const searchParams = Object.fromEntries(event.url.searchParams);
    const safeSearchParams = RecordListOptionsSchema.parse(searchParams);

    let r: ListResult<T>;
    if ((safeSearchParams.perPage ?? 0) < 0) {
        const activities: T[] = await pb.collection(Collection[collection])
            .getFullList<T>(safeSearchParams)
        r = {
            items: activities,
            perPage: -1,
            page: 1,
            totalItems: activities.length,
            totalPages: 1
        }
    } else {
        r = await pb.collection(Collection[collection])
            .getList<T>(safeSearchParams.page, safeSearchParams.perPage, { ...safeSearchParams })
    }
    return r
}

export async function show<T>(event: RequestEvent, collection: Collection) {
    const params = event.params
    const safeParams = RecordIdSchema.parse(params);

    const searchParams = Object.fromEntries(event.url.searchParams);
    const safeSearchParams = RecordOptionsSchema.parse(searchParams);

    const r = await pb.collection(collection.toString())
        .getOne<T>(safeParams.id, safeSearchParams)

    return r
}

export async function create<T>(event: RequestEvent, schema: ZodSchema, collection: Collection) {
    const searchParams = Object.fromEntries(event.url.searchParams);
    const safeSearchParams = RecordOptionsSchema.parse(searchParams);

    const data = await event.request.json();
    const safeData = schema.parse(data);

    const r = await pb.collection(Collection[collection]).create<T>(safeData, {...safeSearchParams, requestKey: null})

    return r
}

export async function update<T>(event: RequestEvent, schema: ZodSchema, collection: Collection) {
    const params = event.params
    const safeParams = RecordIdSchema.parse(params);

    const searchParams = Object.fromEntries(event.url.searchParams);
    const safeSearchParams = RecordOptionsSchema.parse(searchParams);

    const data = await event.request.json();
    const safeData = schema.parse(data);

    const r = await pb.collection(Collection[collection]).update<T>(safeParams.id, safeData, safeSearchParams)

    return r
}

export async function upload<T>(event: RequestEvent, collection: Collection) {
    const params = event.params
    const safeParams = RecordIdSchema.parse(params);

    const data = await event.request.formData();

    const r = await pb.collection(Collection[collection]).update<T>(safeParams.id, data)

    return r
}

export async function remove(event: RequestEvent, collection: Collection) {
    const params = event.params
    const safeParams = RecordIdSchema.parse(params);

    const r = await pb.collection(Collection[collection]).delete(safeParams.id)

    return { 'acknowledged': r }
}

export function handleError(e: any) {
    if (e instanceof ZodError) {
        return error(400, { message: "invalid_params", detail: e.issues } as any)
    } else if (e instanceof ClientResponseError && e.status > 0) {
        return error(e.status as NumericRange<400, 599>, {...e.response, message: e.message, detail: e.originalError.data } as any)
    } else if (e instanceof SyntaxError) {
        return error(400, "invalid_json")
    } else {
        return error(500, e.toString())
    }
}