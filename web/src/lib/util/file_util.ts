import { pb } from "$lib/pocketbase";
import type { FileOptions } from "pocketbase";

export function getFileURL(record: { [key: string]: any; }, filename?: string) {
    if (!filename) {
        return "";
    }

    // return pb.files.getUrl(record, filename);
    return `/api/v1/files/${record.collectionId}/${record.id}/${filename}`
}