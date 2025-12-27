import { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Trash2, Shield, ShieldOff, LogOut, Edit2, Search } from 'lucide-react';
import { Modal } from '@/shared/components/Modal';
import { Button } from '@/shared/components/Button';
import { chatApi } from '../api';
import type { Chat } from '../types';
import type { User } from '@/features/auth/types';
import { useAuthStore } from '@/features/auth/store';
import { clsx } from 'clsx';

interface ChatInfoModalProps {
    isOpen: boolean;
    onClose: () => void;
    chat: Chat;
}

export const ChatInfoModal = ({ isOpen, onClose, chat }: ChatInfoModalProps) => {
    const { user: currentUser } = useAuthStore();
    const queryClient = useQueryClient();

    const [isEditingTitle, setIsEditingTitle] = useState(false);
    const [title, setTitle] = useState(chat.title || chat.name || '');
    const [inviteQuery, setInviteQuery] = useState('');
    const [inviteResults, setInviteResults] = useState<User[]>([]);

    useEffect(() => {
        setTitle(chat.title || chat.name || '');
    }, [chat]);

    // Fetch members
    const { data: members, isLoading, refetch } = useQuery({
        queryKey: ['chatMembers', chat.id],
        queryFn: () => chatApi.getChatMembers(chat.id),
        enabled: isOpen,
    });

    const myMember = members?.find(m => m.user_id === currentUser?.id);
    const isAdmin = myMember?.role === 'owner' || myMember?.role === 'admin';
    const isGroup = chat.type === 2;

    const updateTitleMutation = useMutation({
        mutationFn: (newTitle: string) => chatApi.updateGroupInfo(chat.id, newTitle),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['chats'] });
            setIsEditingTitle(false);
        },
    });

    const inviteMutation = useMutation({
        mutationFn: (userId: number) => chatApi.inviteToChat(chat.id, userId),
        onSuccess: () => {
            refetch();
            setInviteQuery('');
            setInviteResults([]);
        },
    });

    const kickMutation = useMutation({
        mutationFn: (userId: number) => chatApi.kickMember(chat.id, userId),
        onSuccess: () => refetch(),
    });

    const leaveMutation = useMutation({
        mutationFn: () => chatApi.leaveChat(chat.id),
        onSuccess: () => {
            onClose();
            queryClient.invalidateQueries({ queryKey: ['chats'] });
            window.location.reload();
        },
    });

    const promoteMutation = useMutation({
        mutationFn: (userId: number) => chatApi.promoteMember(chat.id, userId),
        onSuccess: () => refetch(),
    });

    const demoteMutation = useMutation({
        mutationFn: (userId: number) => chatApi.demoteMember(chat.id, userId),
        onSuccess: () => refetch(),
    });

    // Search for users to invite
    useEffect(() => {
        if (!inviteQuery.trim()) {
            setInviteResults([]);
            return;
        }
        const timer = setTimeout(async () => {
            try {
                const results = await chatApi.searchUsers(inviteQuery);
                const existingIds = new Set(members?.map(m => m.user_id) || []);
                setInviteResults(results.filter(u => !existingIds.has(u.id)));
            } catch (e) {
                console.error(e);
            }
        }, 300);
        return () => clearTimeout(timer);
    }, [inviteQuery, members]);

    if (!isOpen) return null;

    return (
        <Modal isOpen={isOpen} onClose={onClose} title={isGroup ? 'Group Info' : 'Chat Info'}>
            <div className="space-y-6">
                {/* Header / Title */}
                <div className="flex items-center justify-between">
                    {isEditingTitle ? (
                        <div className="flex gap-2 w-full">
                            <input
                                value={title}
                                onChange={(e) => setTitle(e.target.value)}
                                className="flex-1 bg-background border border-border-subtle rounded px-3 py-1 text-text-primary"
                                autoFocus
                            />
                            <Button onClick={() => updateTitleMutation.mutate(title)} disabled={updateTitleMutation.isPending}>Save</Button>
                            <Button variant="ghost" onClick={() => setIsEditingTitle(false)}>Cancel</Button>
                        </div>
                    ) : (
                        <div className="flex items-center gap-2">
                            <h2 className="text-xl font-semibold text-text-primary">{chat.title || chat.name || 'Untitled'}</h2>
                            {isGroup && isAdmin && (
                                <button onClick={() => setIsEditingTitle(true)} className="text-text-tertiary hover:text-text-secondary">
                                    <Edit2 className="w-4 h-4" />
                                </button>
                            )}
                        </div>
                    )}
                </div>

                {/* Add Member (Group Only + Admin) */}
                {isGroup && (
                    <div className="space-y-2">
                        <h3 className="text-sm font-medium text-text-secondary uppercase">Add Members</h3>
                        <div className="relative">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-tertiary" />
                            <input
                                placeholder="Search users to invite..."
                                value={inviteQuery}
                                onChange={(e) => setInviteQuery(e.target.value)}
                                className="w-full bg-background border border-border-subtle rounded px-10 py-2 text-sm text-text-primary focus:outline-none focus:border-brand-primary"
                            />
                        </div>
                        {inviteResults.length > 0 && (
                            <div className="bg-surface border border-border-subtle rounded mt-1 max-h-40 overflow-y-auto">
                                {inviteResults.map(user => (
                                    <div key={user.id} className="flex items-center justify-between p-2 hover:bg-background/50">
                                        <div className="flex items-center gap-2">
                                            <div className="w-8 h-8 rounded-full bg-brand-primary/10 flex items-center justify-center text-brand-primary text-xs font-medium">
                                                {user.email[0].toUpperCase()}
                                            </div>
                                            <span className="text-sm text-text-primary">{user.email}</span>
                                        </div>
                                        <Button size="sm" variant="ghost" onClick={() => inviteMutation.mutate(user.id)}>Add</Button>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                )}

                {/* Members List */}
                <div className="space-y-2">
                    <h3 className="text-sm font-medium text-text-secondary uppercase">Members ({members?.length || 0})</h3>
                    <div className="space-y-1 max-h-60 overflow-y-auto custom-scrollbar">
                        {isLoading ? (
                            <div className="text-sm text-text-tertiary">Loading members...</div>
                        ) : (
                            members?.map((member) => (
                                <div key={member.user_id} className="flex items-center justify-between p-2 rounded hover:bg-surface/50 transition-colors">
                                    <div className="flex items-center gap-3">
                                        <div className="w-9 h-9 rounded-full bg-gradient-to-br from-brand-primary to-brand-hover flex items-center justify-center text-white text-sm font-medium shadow-sm">
                                            {member.user?.email?.[0].toUpperCase() || '?'}
                                        </div>
                                        <div>
                                            <p className="text-sm font-medium text-text-primary">
                                                {member.user?.email || `User ${member.user_id}`}
                                                {member.user_id === currentUser?.id && <span className="text-text-tertiary ml-1">(You)</span>}
                                            </p>
                                            <span className={clsx(
                                                "text-xs px-1.5 py-0.5 rounded-full inline-block mt-0.5",
                                                (member.role === 'owner' || member.role === 'admin') ? "bg-brand-primary/10 text-brand-primary" : "text-text-tertiary"
                                            )}>
                                                {member.role === 'owner' ? 'Owner' : member.role === 'admin' ? 'Admin' : 'Member'}
                                            </span>
                                        </div>
                                    </div>

                                    {/* Actions */}
                                    {isGroup && isAdmin && member.user_id !== currentUser?.id && (
                                        <div className="flex items-center gap-1">
                                            {member.role === 'member' ? (
                                                <Button size="icon" variant="ghost" title="Promote to Admin" onClick={() => promoteMutation.mutate(member.user_id)}>
                                                    <Shield className="w-4 h-4 text-text-tertiary hover:text-brand-primary" />
                                                </Button>
                                            ) : (
                                                <Button size="icon" variant="ghost" title="Demote to Member" onClick={() => demoteMutation.mutate(member.user_id)}>
                                                    <ShieldOff className="w-4 h-4 text-text-tertiary hover:text-yellow-500" />
                                                </Button>
                                            )}
                                            <Button size="icon" variant="ghost" title="Remove from Group" onClick={() => kickMutation.mutate(member.user_id)} className="hover:bg-red-500/10">
                                                <Trash2 className="w-4 h-4 text-text-tertiary hover:text-red-500" />
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            ))
                        )}
                    </div>
                </div>

                {/* Footer Actions */}
                <div className="pt-4 border-t border-border-subtle flex justify-end">
                    {isGroup && (
                        <Button
                            variant="ghost"
                            className="text-red-500 hover:text-red-600 hover:bg-red-500/10"
                            onClick={() => {
                                if (confirm('Are you sure you want to leave this group?')) {
                                    leaveMutation.mutate();
                                }
                            }}
                        >
                            <LogOut className="w-4 h-4 mr-2" />
                            Leave Group
                        </Button>
                    )}
                </div>
            </div>
        </Modal>
    );
};
