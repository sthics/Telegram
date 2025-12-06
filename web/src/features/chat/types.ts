import type { User } from '@/features/auth/types';

export interface Message {
    id: number;
    chat_id: number;
    user_id: number;
    body: string;
    media_url?: string;
    reply_to_id?: number;
    reactions?: any; // JSONB
    created_at: string; // ISO string
    status?: number; // 1=Sent, 2=Delivered, 3=Read
    user?: User; // Sender details
}

export interface Chat {
    id: number;
    type: number; // 1 = private, 2 = group
    title?: string;
    created_at: string;
    // Computed/Client-side props
    name?: string;
    online?: boolean; // Computed
    lastMessage?: Message;
    unreadCount?: number;
}

export interface CreateChatRequest {
    type: number; // 1 = private, 2 = group
    title?: string;
    memberIds: number[];
}

export interface ChatMember {
    chat_id: number;
    user_id: number;
    role: 'admin' | 'member';
    joined_at: string;
    user?: User;
}
