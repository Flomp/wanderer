import cryptoRandomString from "crypto-random-string";

export type Toast = {
    id?: string;
    type: "info" | "success" | "warning" | "error"
    icon: string;
    text: string;
}

class ToastStore {
    toasts: Toast[] = $state([]);
}

export const toastStore = new ToastStore()


export function show_toast(newToast: Toast, duration: number = 3000) {
    newToast.id = cryptoRandomString({ length: 15 })
    toastStore.toasts.push(newToast);

    setTimeout(() => {
        closeToast(newToast)
    }, duration);
}

export function closeToast(toast: Toast) {
    const toastIndex = toastStore.toasts.findIndex(t => t.id == toast.id)    
    toastStore.toasts.splice(toastIndex, 1)
}