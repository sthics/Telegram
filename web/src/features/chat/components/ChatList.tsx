import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import { useQuery } from '@tanstack/react-query';
import { clsx } from 'clsx';

export const ChatList = () => {
    const { activeChat, setActiveChat } = useChatStore();

    const { data: chats, isLoading } = useQuery({
        queryKey: ['chats'],
        queryFn: chatApi.getChats,
    });

    if (isLoading) {
        return (
            <div className="p-4 text-center text-text-secondary text-sm">
                Loading chats...
            </div>
        );
    }

    if (!chats || chats.length === 0) {
        return (
            <div className="p-4 text-center text-text-secondary text-sm">
                No chats yet. Start a new conversation!
            </div>
        );
    }

    return (
        <div className="flex flex-col">
            {chats.map((chat) => (
                <button
                    key={chat.id}
                    onClick={() => setActiveChat(chat)}
                    className={clsx(
                        'flex items-center p-3 w-full text-left transition-colors duration-200 hover:bg-surface-hover',
                        activeChat?.id === chat.id && 'bg-surface-hover border-l-2 border-brand-primary'
                    )}
                >
                    {/* Avatar Placeholder */}
                    <div className="w-12 h-12 rounded-full bg-brand-primary/20 flex items-center justify-center text-brand-primary font-medium text-lg shrink-0">
                        {chat.name ? chat.name.charAt(0) : 'U'}
                    </div>

                    <div className="ml-3 flex-1 min-w-0">
                        <div className="flex justify-between items-baseline">
                            <h3 className="text-sm font-medium text-text-primary truncate">
                                {chat.name || 'Unknown Chat'}
                            </h3>
                            {chat.lastMessage && (
                                <span className="text-xs text-text-secondary">
                                    {new Date(chat.lastMessage.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                </span>
                            )}
                        </div>
                        <div className="flex justify-between items-center mt-1">
                            <p className="text-sm text-text-secondary truncate pr-2">
                                {chat.lastMessage?.body || 'No messages yet'}
                            </p>
                            {chat.unreadCount !== undefined && chat.unreadCount > 0 && (
                                <span className="bg-brand-primary text-white text-[11px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center">
                                    {chat.unreadCount}
                                </span>
                            )}
                        </div>
                    </div>
                </button>
            ))}
        </div>
    );
};
