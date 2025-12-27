import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Search, Plus, UserPlus, Settings, MessageCircle } from 'lucide-react';
import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import type { User } from '@/features/auth/types';
import { clsx } from 'clsx';
import { Button } from '@/shared/components/Button';
import { Avatar } from '@/shared/components/Avatar';
import { SkeletonChatItem } from '@/shared/components/Skeleton';
import { CreateChatModal } from './CreateChatModal';
import { ProfileSettings } from '@/features/auth/components/ProfileSettings';
import { useAuthStore } from '@/features/auth/store';

export const ChatList = () => {
    const { activeChat, setActiveChat } = useChatStore();
    const currentUser = useAuthStore((state) => state.user);
    const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
    const [isProfileOpen, setIsProfileOpen] = useState(false);
    const [searchQuery, setSearchQuery] = useState('');
    const [debouncedQuery, setDebouncedQuery] = useState('');
    const queryClient = useQueryClient();

    // Debounce search
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedQuery(searchQuery);
        }, 300);
        return () => clearTimeout(timer);
    }, [searchQuery]);

    // Fetch Chats
    const { data: chats, isLoading: isChatsLoading } = useQuery({
        queryKey: ['chats'],
        queryFn: chatApi.getChats,
    });

    // Remote Search for Users
    const { data: userResults } = useQuery({
        queryKey: ['searchUsers', debouncedQuery],
        queryFn: () => chatApi.searchUsers(debouncedQuery),
        enabled: debouncedQuery.length >= 3,
    });

    // Create Chat Mutation
    const createChatMutation = useMutation({
        mutationFn: chatApi.createChat,
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['chats'] });
        },
    });

    const handleUserSelect = async (user: User) => {
        try {
            await createChatMutation.mutateAsync({
                type: 1,
                memberIds: [user.id],
            });
            setSearchQuery('');
        } catch (error) {
            console.error("Failed to start chat", error);
        }
    };

    const filteredChats = useMemo(() => {
        if (!chats) return [];
        if (!debouncedQuery) return chats;
        return chats.filter(c => c.name?.toLowerCase().includes(debouncedQuery.toLowerCase()));
    }, [chats, debouncedQuery]);

    const formatTime = (dateString: string) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffDays = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60 * 24));

        if (diffDays === 0) {
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } else if (diffDays === 1) {
            return 'Yesterday';
        } else if (diffDays < 7) {
            return date.toLocaleDateString([], { weekday: 'short' });
        } else {
            return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
        }
    };

    return (
        <div className="flex flex-col h-full bg-bg-raised">
            {/* Header */}
            <div className="px-4 py-3 border-b border-border-subtle flex items-center justify-between shrink-0">
                <div className="flex items-center gap-3">
                    <Button
                        size="icon-sm"
                        variant="ghost"
                        onClick={() => setIsProfileOpen(true)}
                        title="Settings"
                        className="rounded-full"
                    >
                        <Settings className="w-5 h-5" />
                    </Button>
                    <h1 className="text-h2 text-text-primary">Chats</h1>
                </div>
                <Button
                    size="icon-sm"
                    variant="ghost"
                    onClick={() => setIsCreateModalOpen(true)}
                    className="rounded-full"
                    title="New chat"
                >
                    <Plus className="w-5 h-5" />
                </Button>
            </div>

            {/* Search */}
            <div className="p-3 border-b border-border-subtle shrink-0">
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-tertiary" />
                    <input
                        placeholder="Search chats or users..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-10 pr-4 py-2.5 bg-bg border border-border-default rounded-lg text-body text-text-primary placeholder:text-text-tertiary focus:outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 transition-all"
                    />
                </div>
            </div>

            {/* Chat List */}
            <div className="flex-1 overflow-y-auto">
                {isChatsLoading ? (
                    <div className="divide-y divide-border-subtle">
                        {Array.from({ length: 8 }).map((_, i) => (
                            <SkeletonChatItem key={i} className={`delay-${i * 50}`} />
                        ))}
                    </div>
                ) : (
                    <div>
                        {/* Local Matches */}
                        {filteredChats.length > 0 && (
                            <div>
                                {debouncedQuery && (
                                    <div className="px-4 py-2 text-label text-text-tertiary uppercase tracking-wider bg-bg/50 sticky top-0">
                                        My Chats
                                    </div>
                                )}
                                {filteredChats.map((chat) => (
                                    <div
                                        key={chat.id}
                                        onClick={() => setActiveChat(chat)}
                                        className={clsx(
                                            "flex items-center gap-3 px-4 py-3 cursor-pointer transition-all duration-150",
                                            "hover:bg-bg-elevated",
                                            activeChat?.id === chat.id
                                                ? "bg-brand-500/10 hover:bg-brand-500/15 border-l-2 border-brand-500"
                                                : "border-l-2 border-transparent"
                                        )}
                                    >
                                        <Avatar
                                            name={chat.name || 'Unknown'}
                                            size="lg"
                                            status={chat.type === 1 && chat.online ? 'online' : undefined}
                                            showStatus={chat.type === 1}
                                        />
                                        <div className="flex-1 min-w-0">
                                            <div className="flex justify-between items-baseline gap-2 mb-0.5">
                                                <h3 className={clsx(
                                                    "text-body truncate flex-1",
                                                    chat.unreadCount && chat.unreadCount > 0
                                                        ? "font-semibold text-text-primary"
                                                        : "font-medium text-text-primary",
                                                    activeChat?.id === chat.id && "text-brand-500"
                                                )}>
                                                    {chat.name || 'Unknown Chat'}
                                                </h3>
                                                <span className={clsx(
                                                    "text-caption whitespace-nowrap shrink-0",
                                                    chat.unreadCount && chat.unreadCount > 0
                                                        ? "text-brand-500 font-medium"
                                                        : "text-text-tertiary"
                                                )}>
                                                    {chat.lastMessage ? formatTime(chat.lastMessage.created_at) : ''}
                                                </span>
                                            </div>
                                            <div className="flex justify-between items-center gap-2">
                                                <p className={clsx(
                                                    "text-body-sm truncate flex-1",
                                                    chat.unreadCount && chat.unreadCount > 0
                                                        ? "text-text-primary"
                                                        : "text-text-secondary"
                                                )}>
                                                    {chat.lastMessage ? (
                                                        <>
                                                            {chat.lastMessage.user_id === currentUser?.id && (
                                                                <span className="text-text-tertiary">You: </span>
                                                            )}
                                                            {chat.lastMessage.media_url ? '📷 Photo' : chat.lastMessage.body}
                                                        </>
                                                    ) : (
                                                        <span className="text-text-tertiary italic">No messages yet</span>
                                                    )}
                                                </p>
                                                {chat.unreadCount && chat.unreadCount > 0 ? (
                                                    <span className="bg-brand-500 text-white text-caption font-semibold px-1.5 py-0.5 rounded-full min-w-[20px] text-center shrink-0 animate-scale-in">
                                                        {chat.unreadCount > 99 ? '99+' : chat.unreadCount}
                                                    </span>
                                                ) : null}
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        )}

                        {/* Global Results */}
                        {debouncedQuery.length >= 3 && userResults && userResults.length > 0 && (
                            <div>
                                <div className="px-4 py-2 text-label text-text-tertiary uppercase tracking-wider bg-bg/50 sticky top-0 border-t border-border-subtle">
                                    Global Search
                                </div>
                                {userResults
                                    .filter(u => u.id !== currentUser?.id && !filteredChats.some(c => c.type === 1 && c.name === (u.username || u.email.split('@')[0])))
                                    .map(user => (
                                        <div
                                            key={user.id}
                                            onClick={() => handleUserSelect(user)}
                                            className="flex items-center gap-3 px-4 py-3 cursor-pointer transition-all duration-150 hover:bg-bg-elevated group"
                                        >
                                            <Avatar
                                                name={user.username || user.email}
                                                size="lg"
                                            />
                                            <div className="flex-1 min-w-0">
                                                <h3 className="text-body font-medium text-text-primary truncate">
                                                    {user.username || user.email.split('@')[0]}
                                                </h3>
                                                <p className="text-body-sm text-text-secondary truncate">
                                                    {user.email}
                                                </p>
                                            </div>
                                            <Button
                                                size="icon-sm"
                                                variant="ghost"
                                                className="text-brand-500 opacity-0 group-hover:opacity-100 transition-opacity"
                                            >
                                                <UserPlus className="w-4 h-4" />
                                            </Button>
                                        </div>
                                    ))}
                            </div>
                        )}

                        {/* Empty States */}
                        {!debouncedQuery && filteredChats.length === 0 && (
                            <div className="empty-state text-text-tertiary">
                                <MessageCircle className="w-12 h-12 mb-4 opacity-50" />
                                <p className="text-body font-medium mb-1">No conversations yet</p>
                                <p className="text-body-sm">Start a new chat to begin messaging</p>
                            </div>
                        )}

                        {debouncedQuery && filteredChats.length === 0 && (!userResults || userResults.length === 0) && (
                            <div className="empty-state text-text-tertiary">
                                <Search className="w-10 h-10 mb-3 opacity-50" />
                                <p className="text-body">No results found</p>
                                <p className="text-body-sm mt-1">Try a different search term</p>
                            </div>
                        )}
                    </div>
                )}
            </div>

            <CreateChatModal
                isOpen={isCreateModalOpen}
                onClose={() => setIsCreateModalOpen(false)}
            />

            <ProfileSettings
                isOpen={isProfileOpen}
                onClose={() => setIsProfileOpen(false)}
            />
        </div>
    );
};
