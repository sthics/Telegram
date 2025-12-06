import { useEffect, useRef, useState } from 'react';
import { useAuthStore } from '@/features/auth/store';
import { useQueryClient } from '@tanstack/react-query';
import type { Message } from '@/features/chat/types';

interface WebSocketMessage {
    type: string;
    [key: string]: any;
}

export const useWebSocket = () => {
    const token = useAuthStore((state) => state.token);
    const socketRef = useRef<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const queryClient = useQueryClient();

    useEffect(() => {
        if (!token) return;

        const wsUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8080'}/v1/ws?token=${token}`.replace('http', 'ws');
        const ws = new WebSocket(wsUrl);

        ws.onopen = () => {
            console.log('WebSocket Connected');
            setIsConnected(true);
        };

        ws.onclose = () => {
            console.log('WebSocket Disconnected');
            setIsConnected(false);
            // Simple reconnect logic could go here
        };

        ws.onmessage = (event) => {
            try {
                const data: WebSocketMessage = JSON.parse(event.data);

                if (data.type === 'Message') {
                    const message = data as unknown as Message; // It now matches our Message type
                    // Optimistically update the cache
                    queryClient.setQueryData(['messages', message.chat_id], (old: Message[] | undefined) => {
                        if (!old) return [message];
                        // Avoid duplicates if we already added it locally (though we don't do that yet)
                        if (old.find(m => m.id === message.id)) return old;
                        return [message, ...old]; // Backend returns DESC, but we render reversed... wait.
                        // ChatWindow renders: messages?.slice().reverse()
                        // API returns: newest first (DESC).
                        // So [NewMsg, ...OldMsgs] maintains Newest First order. 
                        // Correct.
                    });

                    // Also refetch chat list to update last message preview
                    queryClient.invalidateQueries({ queryKey: ['chats'] });
                }
            } catch (error) {
                console.error('WebSocket Error:', error);
            }
        };

        socketRef.current = ws;

        return () => {
            ws.close();
        };
    }, [token, queryClient]);

    const sendJson = (data: any) => {
        if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(data));
        } else {
            console.warn('WebSocket not connected');
        }
    };

    return { isConnected, sendJson };
};
