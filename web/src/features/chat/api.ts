import { api } from '@/shared/api/client';
import type { Chat, Message, CreateChatRequest, ChatMember } from './types';
import type { User } from '@/features/auth/types';

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

    markRead: async (chatId: number, lastReadId: number): Promise<void> => {
        await api.post(`/chats/${chatId}/read`, { lastReadId });
    },

    createChat: async (data: CreateChatRequest): Promise<{ chatId: number }> => {
        const response = await api.post<{ chatId: number }>('/chats', data);
        return response.data;
    },

    sendMessage: async (chatId: number, body: string, mediaUrl?: string): Promise<{ messageId: number }> => {
        const response = await api.post<{ messageId: number }>(`/chats/${chatId}/messages`, { body, mediaUrl });
        return response.data;
    },

    getPresignedUrl: async (filename: string, contentType: string): Promise<{ uploadUrl: string; objectKey: string }> => {
        const response = await api.post<{ uploadUrl: string; objectKey: string }>('/uploads/presigned', {
            filename,
            contentType,
        });
        return response.data;
    },

    uploadFileToUrl: async (url: string, file: File, contentType: string): Promise<void> => {
        await fetch(url, {
            method: 'PUT',
            body: file,
            headers: {
                'Content-Type': contentType,
            },
        });
    },

    searchUsers: async (query: string): Promise<User[]> => {
        const response = await api.get<User[]>('/users', {
            params: { q: query },
        });
        return response.data;
    },

    getChatMembers: async (chatId: number): Promise<ChatMember[]> => {
        const response = await api.get<ChatMember[]>(`/chats/${chatId}/members`);
        return response.data;
    },

    updateGroupInfo: async (chatId: number, title: string): Promise<void> => {
        await api.patch(`/chats/${chatId}`, { title });
    },

    inviteToChat: async (chatId: number, userId: number): Promise<void> => {
        await api.post(`/chats/${chatId}/invite`, { userId });
    },

    leaveChat: async (chatId: number): Promise<void> => {
        await api.delete(`/chats/${chatId}/members`);
    },

    kickMember: async (chatId: number, userId: number): Promise<void> => {
        await api.delete(`/chats/${chatId}/members/${userId}`);
    },

    promoteMember: async (chatId: number, userId: number): Promise<void> => {
        await api.post(`/chats/${chatId}/members/${userId}/promote`);
    },

    demoteMember: async (chatId: number, userId: number): Promise<void> => {
        await api.post(`/chats/${chatId}/members/${userId}/demote`);
    },
};
