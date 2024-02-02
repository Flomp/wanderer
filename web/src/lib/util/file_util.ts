import { pb } from "$lib/pocketbase";
import type { FileOptions } from "pocketbase";

export function getFileURL(record: { [key: string]: any; }, filename?: string, options?: FileOptions) {
    if(!filename) {
        return "";
    }
    return pb.files.getUrl(record, filename, options);
}