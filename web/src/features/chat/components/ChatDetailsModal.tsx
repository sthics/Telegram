import { useState, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { X, UserPlus, Shield, ShieldOff, LogOut, Trash2, Edit2, Check } from 'lucide-react';
import { Modal } from '@/shared/components/Modal';
import { Button } from '@/shared/components/Button';
import { chatApi } from '../api';
import { useAuthStore } from '@/features/auth/store';
import type { Chat } from '../types';
import { clsx } from 'clsx';

interface ChatDetailsModalProps {
    isOpen: boolean;
    onClose: () => void;
    chat: Chat;
}

export const ChatDetailsModal = ({ isOpen, onClose, chat }: ChatDetailsModalProps) => {
    const queryClient = useQueryClient();
    const currentUser = useAuthStore((state) => state.user);
    // const [isInviteOpen, setIsInviteOpen] = useState(false); // TODO: Implement InviteModal
    const [isEditingName, setIsEditingName] = useState(false);
    const [newName, setNewName] = useState(chat.name || '');

    // Fetch members
    const { data: members = [] } = useQuery({
        queryKey: ['chatMembers', chat.id],
        queryFn: () => chatApi.getChatMembers(chat.id),
        enabled: isOpen,
    });

    const myRole = useMemo(() => {
        return members.find(m => m.user_id === currentUser?.id)?.role;
    }, [members, currentUser?.id]);

    const isAdmin = myRole === 'owner' || myRole === 'admin';

    // Mutations
    const updateChatMutation = useMutation({
        mutationFn: () => chatApi.updateGroupInfo(chat.id, newName),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['chats'] });
            setIsEditingName(false);
        },
    });

    const promoteMutation = useMutation({
        mutationFn: (userId: number) => chatApi.promoteMember(chat.id, userId),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['chatMembers', chat.id] }),
    });

    const demoteMutation = useMutation({
        mutationFn: (userId: number) => chatApi.demoteMember(chat.id, userId),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['chatMembers', chat.id] }),
    });

    const kickMutation = useMutation({
        mutationFn: (userId: number) => chatApi.kickMember(chat.id, userId),
        onSuccess: () => queryClient.invalidateQueries({ queryKey: ['chatMembers', chat.id] }),
    });

    const leaveMutation = useMutation({
        mutationFn: () => chatApi.leaveChat(chat.id),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['chats'] });
            onClose();
        },
    });

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="Chat Info">
            <div className="space-y-6">
                {/* Header / Name */}
                <div className="flex items-center justify-between gap-4 p-4 bg-app rounded-lg border border-border-subtle">
                    {isEditingName ? (
                        <div className="flex-1 flex gap-2">
                            <input
                                value={newName}
                                onChange={(e) => setNewName(e.target.value)}
                                className="flex-1 px-3 py-1.5 bg-surface border border-border-subtle rounded text-sm focus:outline-none focus:border-brand-primary"
                                autoFocus
                            />
                            <Button size="icon" onClick={() => updateChatMutation.mutate()} disabled={!newName.trim()}>
                                <Check className="w-4 h-4 text-green-500" />
                            </Button>
                            <Button size="icon" variant="ghost" onClick={() => setIsEditingName(false)}>
                                <X className="w-4 h-4" />
                            </Button>
                        </div>
                    ) : (
                        <>
                            <div>
                                <h3 className="font-semibold text-lg text-text-primary">{chat.name}</h3>
                                <p className="text-xs text-text-secondary">{members.length} members</p>
                            </div>
                            {isAdmin && (
                                <Button size="icon" variant="ghost" onClick={() => { setNewName(chat.name || ''); setIsEditingName(true); }}>
                                    <Edit2 className="w-4 h-4 text-text-tertiary hover:text-text-primary" />
                                </Button>
                            )}
                        </>
                    )}
                </div>

                {/* Members List */}
                <div className="space-y-3">
                    <div className="flex items-center justify-between">
                        <h4 className="text-sm font-medium text-text-secondary">Members</h4>
                        {isAdmin && (
                            <Button size="sm" variant="ghost" className="text-brand-primary hover:text-brand-hover" onClick={() => alert("Invite feature coming soon!")}>
                                <UserPlus className="w-4 h-4 mr-1" />
                                Add Member
                            </Button>
                        )}
                    </div>

                    <div className="max-h-[300px] overflow-y-auto custom-scrollbar space-y-1">
                        {members.map((member) => {
                            const isMe = member.user_id === currentUser?.id;
                            const canManage = isAdmin && !isMe;

                            return (
                                <div key={member.user_id} className="flex items-center justify-between p-2 hover:bg-app rounded-lg group transition-colors">
                                    <div className="flex items-center gap-3">
                                        <div className="w-8 h-8 rounded-full bg-brand-primary/10 flex items-center justify-center text-brand-primary text-sm font-medium">
                                            {member.user?.email.charAt(0).toUpperCase()}
                                        </div>
                                        <div>
                                            <div className="flex items-center gap-2">
                                                <span className={clsx("text-sm font-medium", isMe ? "text-brand-primary" : "text-text-primary")}>
                                                    {member.user?.email.split('@')[0]}
                                                    {isMe && " (You)"}
                                                </span>
                                                {(member.role === 'owner' || member.role === 'admin') && (
                                                    <span className="text-[10px] uppercase font-bold text-brand-primary bg-brand-primary/10 px-1.5 py-0.5 rounded">
                                                        {member.role === 'owner' ? 'Owner' : 'Admin'}
                                                    </span>
                                                )}
                                            </div>
                                            <span className="text-xs text-text-secondary">{member.user?.email}</span>
                                        </div>
                                    </div>

                                    {canManage && (
                                        <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                                            {member.role === 'member' ? (
                                                <Button size="icon" variant="ghost" title="Make Admin" onClick={() => promoteMutation.mutate(member.user_id)}>
                                                    <Shield className="w-4 h-4 text-text-tertiary hover:text-green-500" />
                                                </Button>
                                            ) : (
                                                <Button size="icon" variant="ghost" title="Dismiss Admin" onClick={() => demoteMutation.mutate(member.user_id)}>
                                                    <ShieldOff className="w-4 h-4 text-text-tertiary hover:text-orange-500" />
                                                </Button>
                                            )}
                                            <Button size="icon" variant="ghost" title="Remove from Group" onClick={() => kickMutation.mutate(member.user_id)}>
                                                <Trash2 className="w-4 h-4 text-text-tertiary hover:text-red-500" />
                                            </Button>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>
                </div>

                {/* Footer Actions */}
                <div className="pt-4 border-t border-border-subtle">
                    <Button
                        variant="ghost"
                        className="w-full text-red-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-950/20 justify-start"
                        onClick={() => leaveMutation.mutate()}
                    >
                        <LogOut className="w-4 h-4 mr-2" />
                        Leave Group
                    </Button>
                </div>
            </div>
        </Modal>
    );
};
