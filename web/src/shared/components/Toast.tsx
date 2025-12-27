import React, { useEffect, useState } from 'react';
import { createPortal } from 'react-dom';
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { clsx } from 'clsx';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface Toast {
    id: string;
    type: ToastType;
    title: string;
    message?: string;
    duration?: number;
}

interface ToastItemProps {
    toast: Toast;
    onDismiss: (id: string) => void;
}

const icons = {
    success: CheckCircle,
    error: AlertCircle,
    warning: AlertTriangle,
    info: Info,
};

const styles = {
    success: 'bg-success/10 border-success/30 text-success',
    error: 'bg-error/10 border-error/30 text-error',
    warning: 'bg-warning/10 border-warning/30 text-warning',
    info: 'bg-info/10 border-info/30 text-info',
};

const ToastItem: React.FC<ToastItemProps> = ({ toast, onDismiss }) => {
    const [isLeaving, setIsLeaving] = useState(false);
    const [progress, setProgress] = useState(100);
    const duration = toast.duration || 5000;
    const Icon = icons[toast.type];

    useEffect(() => {
        const startTime = Date.now();
        const interval = setInterval(() => {
            const elapsed = Date.now() - startTime;
            const remaining = Math.max(0, 100 - (elapsed / duration) * 100);
            setProgress(remaining);

            if (remaining <= 0) {
                clearInterval(interval);
                handleDismiss();
            }
        }, 50);

        return () => clearInterval(interval);
    }, [duration]);

    const handleDismiss = () => {
        setIsLeaving(true);
        setTimeout(() => onDismiss(toast.id), 150);
    };

    return (
        <div
            className={clsx(
                'relative flex items-start gap-3 p-4 rounded-lg border shadow-lg backdrop-blur-sm',
                'transition-all duration-150 ease-out',
                isLeaving ? 'opacity-0 translate-x-4' : 'opacity-100 translate-x-0 animate-slide-in-right',
                styles[toast.type]
            )}
            role="alert"
        >
            <Icon className="w-5 h-5 shrink-0 mt-0.5" />
            <div className="flex-1 min-w-0">
                <p className="text-body font-medium text-text-primary">{toast.title}</p>
                {toast.message && (
                    <p className="text-body-sm text-text-secondary mt-0.5">{toast.message}</p>
                )}
            </div>
            <button
                onClick={handleDismiss}
                className="shrink-0 p-1 rounded hover:bg-white/10 transition-colors"
            >
                <X className="w-4 h-4 text-text-secondary" />
            </button>

            {/* Progress bar */}
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-white/10 rounded-b-lg overflow-hidden">
                <div
                    className={clsx(
                        'h-full transition-all duration-100 ease-linear',
                        toast.type === 'success' && 'bg-success',
                        toast.type === 'error' && 'bg-error',
                        toast.type === 'warning' && 'bg-warning',
                        toast.type === 'info' && 'bg-info'
                    )}
                    style={{ width: `${progress}%` }}
                />
            </div>
        </div>
    );
};

interface ToastContainerProps {
    toasts: Toast[];
    onDismiss: (id: string) => void;
}

export const ToastContainer: React.FC<ToastContainerProps> = ({ toasts, onDismiss }) => {
    const [mounted, setMounted] = useState(false);

    useEffect(() => {
        setMounted(true);
    }, []);

    if (!mounted) return null;

    return createPortal(
        <div className="fixed top-4 right-4 z-toast flex flex-col gap-2 max-w-sm w-full pointer-events-none">
            {toasts.map((toast) => (
                <div key={toast.id} className="pointer-events-auto">
                    <ToastItem toast={toast} onDismiss={onDismiss} />
                </div>
            ))}
        </div>,
        document.body
    );
};
