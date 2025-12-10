import React, { createContext, useContext, useEffect, useRef, useState } from 'react';
import { useAuthStore } from '@/features/auth/store';
import { useQueryClient } from '@tanstack/react-query';
import type { Message, Chat } from '@/features/chat/types';
import { useNotifications } from '../hooks/useNotifications';
import { useChatStore } from '@/features/chat/stores/chatStore';

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
    const { showNotification, requestPermission } = useNotifications();

    // Request permission on mount (or maybe we should let user do it? For now, request on load is standard for simple apps, but better to request on user interaction. MVP: Request on load)
    useEffect(() => {
        requestPermission();
    }, [requestPermission]);

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

                    // Handle Notifications
                    const activeChat = useChatStore.getState().activeChat;
                    const isHidden = document.hidden;

                    if (isHidden || activeChat?.id !== message.chat_id) {
                        const chats = queryClient.getQueryData<Chat[]>(['chats']);
                        const chat = chats?.find(c => c.id === message.chat_id);
                        const title = chat?.name || 'New Message';
                        const body = message.body || (message.media_url ? '📷 Photo' : 'Sent a message');

                        showNotification(title, {
                            body,
                            icon: '/vite.svg', // Placeholder icon
                            tag: `chat-${message.chat_id}` // Group by chat
                        });
                    }

                    // Update the messages cache with the new message
                    // setQueryData should automatically trigger re-renders in subscribed components
                    queryClient.setQueryData(['messages', message.chat_id], (old: Message[] | undefined) => {
                        console.log('WS: Updating cache for chat', message.chat_id, 'Old size:', old?.length);
                        if (!old) {
                            console.log('WS: No existing cache, creating new with message');
                            return [message];
                        }
                        if (old.find(m => m.id === message.id)) {
                            console.log('WS: Message already exists in cache, skipping');
                            return old;
                        }
                        console.log('WS: Adding message to cache, new size:', old.length + 1);
                        // Return a new array reference to ensure React detects the change
                        return [message, ...old];
                    });

                    // Invalidate chats to update last message preview
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
