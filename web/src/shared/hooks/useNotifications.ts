import { useCallback, useEffect, useState } from 'react';

export const useNotifications = () => {
    const [permission, setPermission] = useState<NotificationPermission>('default');

    useEffect(() => {
        if (!('Notification' in window)) {
            console.warn('This browser does not support desktop notifications');
            return;
        }
        setPermission(Notification.permission);
    }, []);

    const requestPermission = useCallback(async () => {
        if (!('Notification' in window)) return 'denied';

        try {
            const p = await Notification.requestPermission();
            setPermission(p);
            return p;
        } catch (error) {
            console.error('Error requesting notification permission:', error);
            return 'denied';
        }
    }, []);

    const showNotification = useCallback((title: string, options?: NotificationOptions) => {
        if (!('Notification' in window)) return;

        if (permission === 'granted') {
            new Notification(title, options);
        } else if (permission !== 'denied') {
            requestPermission().then((p) => {
                if (p === 'granted') {
                    new Notification(title, options);
                }
            });
        }
    }, [permission, requestPermission]);

    return {
        permission,
        requestPermission,
        showNotification
    };
};
