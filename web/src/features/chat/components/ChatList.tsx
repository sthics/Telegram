import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Search, PenSquare, Loader2 } from 'lucide-react';
import { clsx } from 'clsx';
import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import { Button } from '@/shared/components/Button';
import { CreateChatModal } from './CreateChatModal';

export const ChatList = () => {
    const [isNewChatOpen, setIsNewChatOpen] = useState(false);
    const activeChat = useChatStore((state) => state.activeChat);
    const setActiveChat = useChatStore((state) => state.setActiveChat);

    const { data: chats, isLoading } = useQuery({
        queryKey: ['chats'],
        queryFn: chatApi.getChats,
    });

    return (
        <div className="flex flex-col h-full border-r border-border-subtle bg-surface">
            {/* Header */}
            <div className="h-16 px-4 flex items-center justify-between shrink-0">
                <Button
                    className="flex-1 justify-start pl-2 text-text-secondary hover:text-text-primary bg-app hover:bg-app/80 border border-transparent hover:border-border-subtle transition-all"
                    onClick={() => { }} // Could be global search
                >
                    <Search className="w-4 h-4 mr-2" />
                    <span className="text-sm">Search</span>
                </Button>
                <Button
                    size="icon"
                    variant="ghost"
                    className="ml-2 text-text-secondary hover:text-brand-primary"
                    onClick={() => setIsNewChatOpen(true)}
                >
                    <PenSquare className="w-5 h-5" />
                </Button>
            </div>

            {/* Chat List */}
            <div className="flex-1 overflow-y-auto custom-scrollbar">
                {isLoading ? (
                    <div className="flex justify-center pt-8">
                        <Loader2 className="w-6 h-6 animate-spin text-brand-primary" />
                    </div>
                ) : chats && chats.length > 0 ? (
                    chats.map((chat) => (
                        <div
                            key={chat.id}
                            onClick={() => setActiveChat(chat)}
                            className={clsx(
                                'px-4 py-3 cursor-pointer transition-colors',
                                activeChat?.id === chat.id
                                    ? 'bg-brand-primary/10'
                                    : 'hover:bg-app'
                            )}
                        >
                            <div className="flex items-center gap-3">
                                <div className={clsx(
                                    "w-12 h-12 rounded-full flex items-center justify-center text-lg font-medium shrink-0",
                                    activeChat?.id === chat.id ? "bg-brand-primary text-white" : "bg-brand-primary/20 text-brand-primary"
                                )}>
                                    {chat.name ? chat.name.charAt(0) : 'U'}
                                </div>
                                <div className="flex-1 min-w-0">
                                    <div className="flex justify-between items-baseline mb-1">
                                        <h3 className={clsx(
                                            "text-sm font-medium truncate",
                                            activeChat?.id === chat.id ? "text-brand-primary" : "text-text-primary"
                                        )}>
                                            {chat.name || 'Unknown Chat'}
                                        </h3>
                                        <span className="text-xs text-text-tertiary whitespace-nowrap ml-2">
                                            {chat.lastMessage ? new Date(chat.lastMessage.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                                        </span>
                                    </div>
                                    <p className="text-sm text-text-secondary truncate">
                                        {chat.lastMessage ? chat.lastMessage.body : 'No messages yet'}
                                    </p>
                                </div>
                            </div>
                        </div>
                    ))
                ) : (
                    <div className="text-center pt-10 px-6">
                        <p className="text-text-secondary text-sm">No chats yet.</p>
                        <Button variant="link" onClick={() => setIsNewChatOpen(true)}>Start a new chat</Button>
                    </div>
                )}
            </div>

            <CreateChatModal
                isOpen={isNewChatOpen}
                onClose={() => setIsNewChatOpen(false)}
            />
        </div>
    );
};
