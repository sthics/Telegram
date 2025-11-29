import { create } from 'zustand';
import type { Chat } from '../types';

interface ChatState {
    activeChat: Chat | null;
    setActiveChat: (chat: Chat | null) => void;
}

export const useChatStore = create<ChatState>((set) => ({
    activeChat: null,
    setActiveChat: (chat) => set({ activeChat: chat }),
}));
