import { useState, useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { Search, Loader2, UserPlus } from 'lucide-react';
import { Modal } from '@/shared/components/Modal';
import { Button } from '@/shared/components/Button';
import { chatApi } from '../api';
import type { User } from '@/features/auth/types';
import { useAuthStore } from '@/features/auth/store';
import { useWebSocketContext } from '@/shared/providers/WebSocketProvider';

// Assuming you have queryClient set up in your app. Using simple manual fetch for now inside component or custom hook if prefer.
// But better to use hooks.

interface CreateChatModalProps {
    isOpen: boolean;
    onClose: () => void;
}

export const CreateChatModal = ({ isOpen, onClose }: CreateChatModalProps) => {
    const queryClient = useQueryClient();
    const [mode, setMode] = useState<'private' | 'group'>('private');
    const [query, setQuery] = useState('');
    const [debouncedQuery, setDebouncedQuery] = useState('');
    const [searchResults, setSearchResults] = useState<User[]>([]);
    const [isSearching, setIsSearching] = useState(false);

    // Group states
    const [groupName, setGroupName] = useState('');
    const [selectedUsers, setSelectedUsers] = useState<User[]>([]);

    const currentUser = useAuthStore((state) => state.user);
    const { sendJson } = useWebSocketContext();

    // Reset state on open/mode change
    useEffect(() => {
        if (!isOpen) {
            setQuery('');
            setGroupName('');
            setSelectedUsers([]);
            setMode('private');
        }
    }, [isOpen]);

    // Debounce search
    useEffect(() => {
        const timer = setTimeout(() => {
            setDebouncedQuery(query);
        }, 300);
        return () => clearTimeout(timer);
    }, [query]);

    // Perform search
    useEffect(() => {
        const search = async () => {
            if (debouncedQuery.length < 3) {
                setSearchResults([]);
                return;
            }
            setIsSearching(true);
            try {
                const results = await chatApi.searchUsers(debouncedQuery);
                // Filter out current user and already selected users
                setSearchResults(results.filter(u => u.id !== currentUser?.id));
            } catch (error) {
                console.error("Search failed:", error);
            } finally {
                setIsSearching(false);
            }
        };
        search();
    }, [debouncedQuery, currentUser?.id]);

    const handleCreatePrivateChat = async (userId: number) => {
        try {
            const { chatId } = await chatApi.createChat({
                type: 1,
                memberIds: [userId],
            });
            sendJson({ type: 'Subscribe', chatId });
            await queryClient.invalidateQueries({ queryKey: ['chats'] });
            onClose();
        } catch (error) {
            console.error("Failed to create chat:", error);
        }
    };

    const handleCreateGroupChat = async () => {
        if (!groupName.trim() || selectedUsers.length === 0) return;
        try {
            const { chatId } = await chatApi.createChat({
                type: 2,
                title: groupName,
                memberIds: selectedUsers.map(u => u.id),
            });
            sendJson({ type: 'Subscribe', chatId });
            onClose();
        } catch (error) {
            console.error("Failed to create group:", error);
        }
    };

    const toggleUserSelection = (user: User) => {
        if (selectedUsers.find(u => u.id === user.id)) {
            setSelectedUsers(selectedUsers.filter(u => u.id !== user.id));
        } else {
            setSelectedUsers([...selectedUsers, user]);
        }
        setQuery(''); // Optional: clear search after select? strictly no, better to keep searching
    };

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={mode === 'group' ? "New Group" : "New Chat"}>
            <div className="space-y-4">
                {/* Mode Switcher */}
                <div className="flex gap-2 p-1 bg-surface rounded-lg border border-border-subtle">
                    <button
                        className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${mode === 'private' ? 'bg-brand-primary text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'}`}
                        onClick={() => setMode('private')}
                    >
                        Private Chat
                    </button>
                    <button
                        className={`flex-1 py-1.5 text-sm font-medium rounded-md transition-colors ${mode === 'group' ? 'bg-brand-primary text-white shadow-sm' : 'text-text-secondary hover:text-text-primary'}`}
                        onClick={() => setMode('group')}
                    >
                        New Group
                    </button>
                </div>

                {/* Group Name Input */}
                {mode === 'group' && (
                    <div>
                        <input
                            placeholder="Group Name"
                            value={groupName}
                            onChange={(e) => setGroupName(e.target.value)}
                            className="w-full px-4 py-2 bg-app border border-border-subtle rounded-lg text-sm text-text-primary focus:outline-none focus:border-brand-primary"
                        />
                    </div>
                )}

                {/* Selected Users Chips (Group Mode) */}
                {mode === 'group' && selectedUsers.length > 0 && (
                    <div className="flex flex-wrap gap-2">
                        {selectedUsers.map(u => (
                            <div key={u.id} className="flex items-center gap-1 pl-2 pr-1 py-1 bg-brand-primary/10 text-brand-primary rounded-full text-xs font-medium">
                                {u.email}
                                <button onClick={() => toggleUserSelection(u)} className="p-0.5 hover:bg-brand-primary/20 rounded-full">
                                    <div className="w-3 h-3 flex items-center justify-center">×</div>
                                </button>
                            </div>
                        ))}
                    </div>
                )}

                {/* Search Input */}
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-tertiary" />
                    <input
                        type="text"
                        placeholder={mode === 'group' ? "Add members..." : "Search users by email..."}
                        className="w-full pl-10 pr-4 py-2 bg-app border border-border-subtle rounded-lg text-sm text-text-primary focus:outline-none focus:border-brand-primary transition-colors"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        autoFocus
                    />
                </div>

                {/* Results List */}
                <div className="min-h-[200px] max-h-[300px] overflow-y-auto custom-scrollbar">
                    {isSearching ? (
                        <div className="flex justify-center py-8">
                            <Loader2 className="w-6 h-6 animate-spin text-brand-primary" />
                        </div>
                    ) : (
                        <div className="space-y-2">
                            {searchResults.length > 0 ? (
                                searchResults.map((user) => {
                                    const isSelected = selectedUsers.some(u => u.id === user.id);
                                    return (
                                        <div
                                            key={user.id}
                                            className={`flex items-center justify-between p-3 rounded-lg transition-colors cursor-pointer group ${isSelected ? 'bg-brand-primary/5 border border-brand-primary/10' : 'hover:bg-app'}`}
                                            onClick={() => mode === 'group' ? toggleUserSelection(user) : handleCreatePrivateChat(user.id)}
                                        >
                                            <div className="flex items-center gap-3">
                                                <div className="w-10 h-10 rounded-full bg-brand-primary/10 flex items-center justify-center text-brand-primary font-medium">
                                                    {user.username ? user.username.charAt(0).toUpperCase() : user.email.charAt(0).toUpperCase()}
                                                </div>
                                                <div>
                                                    <p className="text-sm font-medium text-text-primary">
                                                        {user.username || user.email.split('@')[0]}
                                                    </p>
                                                    <p className="text-xs text-text-secondary">
                                                        {user.email}
                                                    </p>
                                                </div>
                                            </div>
                                            {mode === 'private' ? (
                                                <Button size="icon" variant="ghost" className="opacity-0 group-hover:opacity-100 transition-opacity">
                                                    <UserPlus className="w-5 h-5 text-brand-primary" />
                                                </Button>
                                            ) : (
                                                <div className={`w-5 h-5 rounded border flex items-center justify-center ${isSelected ? 'bg-brand-primary border-brand-primary' : 'border-border-subtle'}`}>
                                                    {isSelected && <div className="w-2.5 h-2.5 bg-white rounded-sm" />}
                                                </div>
                                            )}
                                        </div>
                                    );
                                })
                            ) : (
                                query.length >= 3 && (
                                    <div className="text-center py-8 text-text-secondary text-sm">
                                        No users found.
                                    </div>
                                )
                            )}
                            {query.length < 3 && (
                                <div className="text-center py-8 text-text-secondary text-sm">
                                    Type at least 3 characters to search.
                                </div>
                            )}
                        </div>
                    )}
                </div>

                {/* Create Group Button */}
                {mode === 'group' && (
                    <div className="pt-2">
                        <Button
                            className="w-full bg-brand-primary hover:bg-brand-hover text-white"
                            disabled={!groupName.trim() || selectedUsers.length === 0}
                            onClick={handleCreateGroupChat}
                        >
                            Create Group ({selectedUsers.length} members)
                        </Button>
                    </div>
                )}
            </div>
        </Modal>
    );
};
