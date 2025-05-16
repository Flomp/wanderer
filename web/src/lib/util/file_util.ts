
export function getFileURL(record: { [key: string]: any; }, filename?: string) {
    if (!filename) {
        return "";
    }
    if (isURL(filename)) {
        return filename;
    }

    return `/api/v1/files/${record.collectionId}/${record.id}/${filename}`
}

export function isURL(value: string) {
    let url
    try {
        url = new URL(value);
    } catch (_) {
        return false;
    }

    return url.protocol === "http:" || url.protocol === "https:";
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

export function isVideoURL(url: string) {
    if (url.startsWith("data")) {
        return url.startsWith("data:video")
    }
    return url.includes("mp4") || url.includes("ogg") || url.includes("webm")
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

function buildFormData(formData: FormData, data: any, parentKey?: string, exclude?: string[]) {
    if (data && typeof data === 'object' && !(data instanceof Date) && !(data instanceof File) && !(data instanceof Blob)) {
        Object.keys(data).forEach(key => {
            if (exclude?.includes(key)) {
                return;
            }
            buildFormData(formData, data[key], parentKey ? `${parentKey}` : key);
        });
    } else {
        const value = data == null ? '' : data;

        formData.append(parentKey!, value);
    }
}

export function objectToFormData(data: Object, exclude?: string[]) {
    const formData = new FormData();

    buildFormData(formData, data, undefined, exclude);

    return formData;
}