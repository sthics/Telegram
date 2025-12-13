import { useState, useMemo, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Search, Plus, Loader2, UserPlus, Settings } from 'lucide-react';
import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import type { User } from '@/features/auth/types';
import { clsx } from 'clsx';
import { Button } from '@/shared/components/Button';
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

    // Combined items to render?
    // We want to show "My Chats" results AND "Global Search" results.

    return (
        <div className="w-80 border-r border-border-subtle flex flex-col h-full bg-surface">
            {/* Header */}
            <div className="p-4 border-b border-border-subtle flex items-center justify-between shrink-0">
                <div className="flex items-center gap-2">
                    <Button size="icon" variant="ghost" onClick={() => setIsProfileOpen(true)} title="Settings">
                        <Settings className="w-5 h-5" />
                    </Button>
                    <h1 className="font-bold text-xl text-text-primary">Chats</h1>
                </div>
                <Button size="icon" variant="ghost" onClick={() => setIsCreateModalOpen(true)}>
                    <Plus className="w-5 h-5" />
                </Button>
            </div>

            {/* Search */}
            <div className="p-3 border-b border-border-subtle bg-app shrink-0">
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-tertiary" />
                    <input
                        placeholder="Search chats..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-full pl-9 pr-4 py-2 bg-surface text-text-primary border border-border-subtle rounded-lg text-sm focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary transition-all placeholder:text-text-tertiary"
                    />
                </div>
            </div>

            {/* Chat List */}
            <div className="flex-1 overflow-y-auto custom-scrollbar">
                {isChatsLoading ? (
                    <div className="flex justify-center pt-8">
                        <Loader2 className="w-6 h-6 animate-spin text-brand-primary" />
                    </div>
                ) : (
                    <div className="divide-y divide-border-subtle">
                        {/* Local Matches */}
                        {filteredChats.length > 0 && (
                            <div>
                                {debouncedQuery && <div className="px-4 py-2 text-xs font-semibold text-text-tertiary uppercase tracking-wider bg-app/50">My Chats</div>}
                                {filteredChats.map((chat) => (
                                    <div
                                        key={chat.id}
                                        onClick={() => setActiveChat(chat)}
                                        className={clsx(
                                            "flex items-center gap-3 p-3 cursor-pointer transition-colors hover:bg-app",
                                            activeChat?.id === chat.id ? "bg-brand-primary/10 hover:bg-brand-primary/15" : ""
                                        )}
                                    >
                                        <div className="relative shrink-0">
                                            <div className="w-12 h-12 rounded-full bg-gradient-to-br from-brand-primary to-brand-hover flex items-center justify-center text-white font-medium shadow-sm">
                                                {chat.name ? chat.name.charAt(0) : 'U'}
                                            </div>
                                            {chat.type === 1 && chat.online && (
                                                <span className="absolute bottom-0 right-0 w-3 h-3 bg-green-500 border-2 border-surface rounded-full"></span>
                                            )}
                                        </div>
                                        <div className="flex-1 min-w-0">
                                            <div className="flex justify-between items-baseline mb-0.5">
                                                <h3 className={clsx(
                                                    "text-sm truncate flex-1 pr-2",
                                                    chat.unreadCount && chat.unreadCount > 0 ? "font-bold text-text-primary" : "font-medium text-text-primary",
                                                    activeChat?.id === chat.id && "!text-brand-primary"
                                                )}>
                                                    {chat.name || 'Unknown Chat'}
                                                </h3>
                                                <span className={clsx(
                                                    "text-xs whitespace-nowrap shrink-0",
                                                    chat.unreadCount && chat.unreadCount > 0 ? "text-brand-primary font-medium" : "text-text-tertiary"
                                                )}>
                                                    {chat.lastMessage ? new Date(chat.lastMessage.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }) : ''}
                                                </span>
                                            </div>
                                            <div className="flex justify-between items-center">
                                                <p className={clsx(
                                                    "text-sm truncate flex-1",
                                                    chat.unreadCount && chat.unreadCount > 0 ? "text-text-primary font-medium" : "text-text-secondary"
                                                )}>
                                                    {chat.lastMessage ? (
                                                        <>
                                                            {chat.lastMessage.user_id === currentUser?.id && <span className="text-text-tertiary">You: </span>}
                                                            {chat.lastMessage.media_url ? '📷 Photo' : chat.lastMessage.body}
                                                        </>
                                                    ) : (
                                                        'No messages yet'
                                                    )}
                                                </p>
                                                {chat.unreadCount && chat.unreadCount > 0 ? (
                                                    <span className="bg-brand-primary text-white text-[10px] font-bold px-1.5 py-0.5 rounded-full min-w-[18px] text-center shrink-0 shadow-sm animate-in zoom-in duration-200">
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
                                <div className="px-4 py-2 text-xs font-semibold text-text-tertiary uppercase tracking-wider bg-app/50 border-t border-border-subtle">Global Search</div>
                                {userResults
                                    .filter(u => u.id !== currentUser?.id && !filteredChats.some(c => c.type === 1 && c.name === (u.username || u.email.split('@')[0])))
                                    .map(user => (
                                        <div
                                            key={user.id}
                                            onClick={() => handleUserSelect(user)}
                                            className="flex items-center gap-3 p-3 cursor-pointer transition-colors hover:bg-app group"
                                        >
                                            <div className="w-12 h-12 rounded-full bg-text-secondary/10 flex items-center justify-center text-text-secondary font-medium group-hover:bg-brand-primary/20 group-hover:text-brand-primary transition-colors">
                                                {user.username ? user.username.charAt(0).toUpperCase() : user.email.charAt(0).toUpperCase()}
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <h3 className="text-sm font-medium text-text-primary truncate">
                                                    {user.username || user.email.split('@')[0]}
                                                </h3>
                                                <p className="text-xs text-text-secondary truncate">
                                                    {user.email}
                                                </p>
                                            </div>
                                            <Button size="icon" variant="ghost" className="text-brand-primary opacity-0 group-hover:opacity-100 transition-opacity">
                                                <UserPlus className="w-4 h-4" />
                                            </Button>
                                        </div>
                                    ))}
                            </div>
                        )}

                        {/* No Results */}
                        {debouncedQuery && filteredChats.length === 0 && (!userResults || userResults.length === 0) && (
                            <div className="p-8 text-center text-text-tertiary">
                                <p className="text-sm">No results found</p>
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
