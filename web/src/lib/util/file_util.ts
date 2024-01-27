import { pb } from "$lib/constants";
import type { FileOptions } from "pocketbase";

export function getFileURL(record: { [key: string]: any; }, filename: string, options?: FileOptions) {
    return pb.files.getUrl(record, filename, options);
}