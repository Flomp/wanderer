
export function getFileURL(record: { [key: string]: any; }, filename?: string) {
    if (!filename) {
        return "";
    }

    return `/api/v1/files/${record.collectionId}/${record.id}/${filename}`
}