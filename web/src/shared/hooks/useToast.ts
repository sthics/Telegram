import { useState, useCallback } from 'react';
import type { Toast, ToastType } from '../components/Toast';

let toastId = 0;

export function useToast() {
    const [toasts, setToasts] = useState<Toast[]>([]);

    const addToast = useCallback((
        type: ToastType,
        title: string,
        message?: string,
        duration?: number
    ) => {
        const id = `toast-${++toastId}`;
        const toast: Toast = { id, type, title, message, duration };
        setToasts((prev) => [...prev, toast]);
        return id;
    }, []);

    const dismissToast = useCallback((id: string) => {
        setToasts((prev) => prev.filter((t) => t.id !== id));
    }, []);

    const success = useCallback((title: string, message?: string) => {
        return addToast('success', title, message);
    }, [addToast]);

    const error = useCallback((title: string, message?: string) => {
        return addToast('error', title, message, 7000);
    }, [addToast]);

    const warning = useCallback((title: string, message?: string) => {
        return addToast('warning', title, message);
    }, [addToast]);

    const info = useCallback((title: string, message?: string) => {
        return addToast('info', title, message);
    }, [addToast]);

    return {
        toasts,
        addToast,
        dismissToast,
        success,
        error,
        warning,
        info,
    };
}
