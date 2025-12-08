import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useAuthStore } from '@/features/auth/store';
import { useQueryClient } from '@tanstack/react-query';
import type { Message } from '@/features/chat/types';

interface WebSocketContextType {
    isConnected: boolean;
    sendJson: (data: any) => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const WebSocketProvider = ({ children }: { children: React.ReactNode }) => {
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
        };

        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);

                if (data.type === 'Message') {
                    const message = data as unknown as Message;
                    console.log('WS: Received message:', message);
                    queryClient.setQueryData(['messages', message.chat_id], (old: Message[] | undefined) => {
                        console.log('WS: Updating cache for chat', message.chat_id, 'Old size:', old?.length);
                        if (!old) return [message];
                        if (old.find(m => m.id === message.id)) return old;
                        return [message, ...old];
                    });
                    queryClient.invalidateQueries({ queryKey: ['chats'] });
                } else if (data.type === 'ReadReceipt') {
                    const { chatId, msgId } = data;
                    console.log('WS: Read receipt for chat', chatId, 'up to', msgId);

                    queryClient.setQueryData(['messages', chatId], (old: Message[] | undefined) => {
                        if (!old) return old;
                        return old.map(msg => {
                            // If message is older than or equal to the read receipt and not yet read
                            if (msg.id <= msgId && msg.status !== 3) {
                                return { ...msg, status: 3 };
                            }
                            return msg;
                        });
                    });
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
        }
    };

    return (
        <WebSocketContext.Provider value={{ isConnected, sendJson }}>
            {children}
        </WebSocketContext.Provider>
    );
};

export const useWebSocketContext = () => {
    const context = useContext(WebSocketContext);
    if (!context) throw new Error('useWebSocketContext must be used within WebSocketProvider');
    return context;
};
