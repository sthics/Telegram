import { useState } from 'react';
import { clsx } from 'clsx';
import { Check, CheckCheck, CornerUpLeft, MessageSquare, SmilePlus } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Message, Reaction } from '../types';
import { useAuthStore } from '@/features/auth/store';
import { chatApi } from '../api';

interface MessageBubbleProps {
    message: Message;
    innerRef?: (node?: Element | null) => void;
    onReply?: (message: Message) => void;
    onViewThread?: (message: Message) => void;
}

const QUICK_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '👎', '🎉', '🔥'];

// Group reactions by emoji with user counts
const groupReactions = (reactions: Reaction[] | null | undefined) => {
    const grouped: Record<string, { count: number; userIds: number[] }> = {};
    if (!reactions || !Array.isArray(reactions)) {
        return grouped;
    }
    reactions.forEach((r) => {
        if (!grouped[r.emoji]) {
            grouped[r.emoji] = { count: 0, userIds: [] };
        }
        grouped[r.emoji].count++;
        grouped[r.emoji].userIds.push(r.user_id);
    });
    return grouped;
};

export const MessageBubble = ({ message, innerRef, onReply, onViewThread }: MessageBubbleProps) => {
    const currentUser = useAuthStore((state) => state.user);
    const isMyMessage = message.user_id === currentUser?.id;
    const queryClient = useQueryClient();
    const [showActions, setShowActions] = useState(false);
    const [showEmojiPicker, setShowEmojiPicker] = useState(false);

    const groupedReactions = groupReactions(message.reactions);

    // Add reaction mutation
    const addReactionMutation = useMutation({
        mutationFn: (emoji: string) => chatApi.addReaction(message.chat_id, message.id, emoji),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['messages', message.chat_id] });
        },
    });

    // Remove reaction mutation
    const removeReactionMutation = useMutation({
        mutationFn: (emoji: string) => chatApi.removeReaction(message.chat_id, message.id, emoji),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['messages', message.chat_id] });
        },
    });

    const handleEmojiSelect = (emoji: string) => {
        const userReacted = groupedReactions[emoji]?.userIds.includes(currentUser?.id ?? -1);
        if (userReacted) {
            removeReactionMutation.mutate(emoji);
        } else {
            addReactionMutation.mutate(emoji);
        }
        setShowEmojiPicker(false);
    };

    const handleReactionClick = (emoji: string) => {
        const userReacted = groupedReactions[emoji]?.userIds.includes(currentUser?.id ?? -1);
        if (userReacted) {
            removeReactionMutation.mutate(emoji);
        } else {
            addReactionMutation.mutate(emoji);
        }
    };

    return (
        <div
            ref={innerRef}
            className={clsx('flex group', isMyMessage ? 'justify-end' : 'justify-start')}
            data-message-id={message.id}
            onMouseEnter={() => setShowActions(true)}
            onMouseLeave={() => { setShowActions(false); setShowEmojiPicker(false); }}
        >
            <div className="flex flex-col max-w-[65%] relative">
                {/* Reply indicator */}
                {message.reply_to_id && (
                    <div className={clsx(
                        "text-xs px-2 py-1 mb-1 rounded border-l-2",
                        isMyMessage
                            ? "bg-white/10 border-white/30 text-white/70 self-end"
                            : "bg-surface-hover border-brand-primary/50 text-text-secondary self-start"
                    )}>
                        <CornerUpLeft className="w-3 h-3 inline mr-1" />
                        Replying to message
                    </div>
                )}

                {/* Message bubble */}
                <div
                    className={clsx(
                        'px-4 py-2 shadow-sm',
                        isMyMessage
                            ? 'bg-brand-primary text-white rounded-lg rounded-tr-none self-end'
                            : 'bg-surface text-text-primary rounded-lg rounded-tl-none self-start'
                    )}
                >
                    {message.media_url && (
                        <div className="mb-2">
                            <img
                                src={message.media_url}
                                alt="attachment"
                                className="rounded-md max-w-full h-auto object-cover max-h-64"
                            />
                        </div>
                    )}
                    {message.body && <p className="text-sm whitespace-pre-wrap">{message.body}</p>}
                    <div className={clsx(
                        "flex items-center justify-end gap-1 mt-1",
                        isMyMessage ? "text-white/70" : "text-text-secondary"
                    )}>
                        <span className="text-[11px]">
                            {new Date(message.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                        {isMyMessage && (
                            message.status === 3 ?
                                <CheckCheck className="w-3 h-3" /> :
                                <Check className="w-3 h-3" />
                        )}
                    </div>
                </div>

                {/* Reactions badge - below the message */}
                {Object.keys(groupedReactions).length > 0 && (
                    <div className={clsx(
                        "flex gap-1 mt-1 bg-surface/80 backdrop-blur-sm rounded-full shadow-sm border border-border-subtle px-2 py-0.5",
                        isMyMessage ? "self-end" : "self-start"
                    )}>
                        {Object.entries(groupedReactions).map(([emoji, { count, userIds }]) => {
                            const userReacted = userIds.includes(currentUser?.id ?? -1);
                            return (
                                <button
                                    key={emoji}
                                    onClick={() => handleReactionClick(emoji)}
                                    className={clsx(
                                        "text-sm flex items-center gap-0.5 transition-colors",
                                        userReacted ? "text-brand-primary" : "text-text-secondary hover:text-brand-primary"
                                    )}
                                >
                                    <span>{emoji}</span>
                                    {count > 1 && <span className="text-xs">{count}</span>}
                                </button>
                            );
                        })}
                    </div>
                )}

                {/* Thread count */}
                {message.reply_count && message.reply_count > 0 && onViewThread && (
                    <button
                        onClick={() => onViewThread(message)}
                        className={clsx(
                            "flex items-center gap-1 mt-1 text-xs text-brand-primary hover:underline",
                            isMyMessage ? "self-end" : "self-start"
                        )}
                    >
                        <MessageSquare className="w-3 h-3" />
                        {message.reply_count} {message.reply_count === 1 ? 'reply' : 'replies'}
                    </button>
                )}

                {/* Action bar - absolute position to prevent layout shift */}
                {showActions && (
                    <div className={clsx(
                        "absolute top-0 flex items-center gap-1 bg-surface rounded-lg shadow-md border border-border-subtle p-1 z-10",
                        isMyMessage ? "right-full mr-2" : "left-full ml-2"
                    )}>
                        <button
                            onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                            className="p-1.5 rounded hover:bg-surface-hover text-text-secondary hover:text-text-primary transition-colors"
                            title="Add reaction"
                        >
                            <SmilePlus className="w-4 h-4" />
                        </button>
                        {onReply && (
                            <button
                                onClick={() => onReply(message)}
                                className="p-1.5 rounded hover:bg-surface-hover text-text-secondary hover:text-text-primary transition-colors"
                                title="Reply"
                            >
                                <CornerUpLeft className="w-4 h-4" />
                            </button>
                        )}
                    </div>
                )}

                {/* Emoji picker dropdown - absolute position */}
                {showEmojiPicker && (
                    <div className={clsx(
                        "absolute top-8 flex flex-wrap gap-1 bg-surface rounded-lg shadow-lg border border-border-subtle p-2 max-w-[200px] z-20",
                        isMyMessage ? "right-full mr-2" : "left-full ml-2"
                    )}>
                        {QUICK_EMOJIS.map((emoji) => (
                            <button
                                key={emoji}
                                onClick={() => handleEmojiSelect(emoji)}
                                className="p-1.5 hover:bg-surface-hover rounded transition-colors text-lg"
                            >
                                {emoji}
                            </button>
                        ))}
                    </div>
                )}
            </div>
        </div>
    );
};
