import { api } from '@/shared/api/client';
import type { Chat, Message, CreateChatRequest } from './types';

export const chatApi = {
    getChats: async (): Promise<Chat[]> => {
        const response = await api.get<Chat[]>('/chats');
        return response.data;
    },

    getMessages: async (chatId: number, limit = 50): Promise<Message[]> => {
        const response = await api.get<Message[]>(`/chats/${chatId}/messages`, {
            params: { limit },
        });
        return response.data;
    },

    createChat: async (data: CreateChatRequest): Promise<{ chatId: number }> => {
        const response = await api.post<{ chatId: number }>('/chats', data);
        return response.data;
    },
};
