
export function getFileURL(record: { [key: string]: any; }, filename?: string) {
    if (!filename) {
        return "";
    }

    return `/api/v1/files/${record.collectionId}/${record.id}/${filename}`
}

export function readAsDataURLAsync(file: File) {
    return new Promise<string>((resolve, reject) => {
        var fr = new FileReader();
        fr.onload = () => {
            resolve(fr.result as string);
        };
        fr.onerror = reject;
        fr.readAsDataURL(file);
    });
}

export function saveAs(data: Blob, fileName: string) {
    var a = document.createElement("a") as HTMLAnchorElement;
    a.setAttribute("style", "display: none");
    document.body.appendChild(a);
    const url = window.URL.createObjectURL(data);
    a.href = url;
    a.download = fileName;
    a.click();
    window.URL.revokeObjectURL(url);
};