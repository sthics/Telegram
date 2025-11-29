import type { User } from '@/features/auth/types';

export interface Message {
    id: number;
    chatId: number;
    userId: number;
    body: string;
    createdAt: string; // ISO string
    user?: User; // Sender details
}

export interface Chat {
    id: number;
    type: 'private' | 'group';
    name: string;
    participants: User[];
    lastMessage?: Message;
    unreadCount?: number;
    createdAt: string;
}

export interface CreateChatRequest {
    type: 'private' | 'group';
    name?: string; // Required for group
    participantIds: number[];
}
