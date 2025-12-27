import { useState } from 'react';
import { clsx } from 'clsx';
import { Check, CheckCheck, CornerUpLeft, MessageSquare, SmilePlus, AlertCircle } from 'lucide-react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { Message, Reaction } from '../types';
import { useAuthStore } from '@/features/auth/store';
import { chatApi } from '../api';
import { Avatar } from '@/shared/components/Avatar';

interface MessageBubbleProps {
    message: Message;
    innerRef?: (node?: Element | null) => void;
    onReply?: (message: Message) => void;
    onViewThread?: (message: Message) => void;
    showAvatar?: boolean;
    isFirstInGroup?: boolean;
    isLastInGroup?: boolean;
    senderName?: string;
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

// Message status component
const MessageStatus = ({ status, isMyMessage }: { status: number; isMyMessage: boolean }) => {
    if (!isMyMessage) return null;

    const statusConfig = {
        1: { icon: Check, label: 'Sending', className: 'text-current opacity-50' },
        2: { icon: Check, label: 'Sent', className: 'text-current opacity-70' },
        3: { icon: CheckCheck, label: 'Read', className: 'text-brand-300' },
    };

    const config = statusConfig[status as keyof typeof statusConfig] || statusConfig[1];
    const Icon = config.icon;

    return (
        <Icon
            className={clsx('w-3.5 h-3.5', config.className)}
            aria-label={config.label}
        />
    );
};

export const MessageBubble = ({
    message,
    innerRef,
    onReply,
    onViewThread,
    showAvatar = true,
    isFirstInGroup = true,
    isLastInGroup = true,
    senderName,
}: MessageBubbleProps) => {
    const currentUser = useAuthStore((state) => state.user);
    const isMyMessage = message.user_id === currentUser?.id;
    const queryClient = useQueryClient();
    const [showActions, setShowActions] = useState(false);
    const [showEmojiPicker, setShowEmojiPicker] = useState(false);

    const groupedReactions = groupReactions(message.reactions);
    const isFailed = message.status === 0;
    const isSending = message.status === 1;

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

    // Bubble shape based on position in group
    const getBubbleRadius = () => {
        if (isMyMessage) {
            if (isFirstInGroup && isLastInGroup) return 'rounded-2xl rounded-br-md';
            if (isFirstInGroup) return 'rounded-2xl rounded-br-md rounded-tr-2xl';
            if (isLastInGroup) return 'rounded-2xl rounded-br-md rounded-tr-lg';
            return 'rounded-2xl rounded-r-lg';
        } else {
            if (isFirstInGroup && isLastInGroup) return 'rounded-2xl rounded-bl-md';
            if (isFirstInGroup) return 'rounded-2xl rounded-bl-md rounded-tl-2xl';
            if (isLastInGroup) return 'rounded-2xl rounded-bl-md rounded-tl-lg';
            return 'rounded-2xl rounded-l-lg';
        }
    };

    return (
        <div
            ref={innerRef}
            className={clsx(
                'flex group gap-2',
                isMyMessage ? 'justify-end' : 'justify-start',
                !isLastInGroup && 'mb-0.5',
                isLastInGroup && 'mb-1'
            )}
            data-message-id={message.id}
            onMouseEnter={() => setShowActions(true)}
            onMouseLeave={() => { setShowActions(false); setShowEmojiPicker(false); }}
        >
            {/* Avatar for other users */}
            {!isMyMessage && (
                <div className="w-8 shrink-0">
                    {showAvatar && isLastInGroup && (
                        <Avatar
                            name={senderName || 'User'}
                            size="sm"
                        />
                    )}
                </div>
            )}

            <div className={clsx(
                'flex flex-col max-w-[70%] relative',
                isMyMessage ? 'items-end' : 'items-start'
            )}>
                {/* Sender name for group chats */}
                {!isMyMessage && isFirstInGroup && senderName && (
                    <span className="text-caption font-medium text-brand-400 mb-1 ml-1">
                        {senderName}
                    </span>
                )}

                {/* Reply indicator */}
                {message.reply_to_id && (
                    <div className={clsx(
                        "text-caption px-3 py-1.5 mb-1 rounded-lg border-l-2 max-w-full",
                        isMyMessage
                            ? "bg-white/10 border-white/40 text-white/80"
                            : "bg-bg-elevated border-brand-500/50 text-text-secondary"
                    )}>
                        <CornerUpLeft className="w-3 h-3 inline mr-1.5 opacity-70" />
                        <span className="truncate">Replying to message</span>
                    </div>
                )}

                {/* Message bubble */}
                <div
                    className={clsx(
                        'px-3.5 py-2 shadow-sm transition-all duration-150',
                        getBubbleRadius(),
                        isMyMessage
                            ? 'bg-brand-500 text-white'
                            : 'bg-bg-raised text-text-primary border border-border-subtle',
                        isSending && 'opacity-70',
                        isFailed && 'bg-error/80'
                    )}
                >
                    {/* Image attachment */}
                    {message.media_url && (
                        <div className="mb-2 -mx-1 -mt-0.5">
                            <img
                                src={message.media_url}
                                alt="attachment"
                                className="rounded-lg max-w-full h-auto object-cover max-h-64"
                                loading="lazy"
                            />
                        </div>
                    )}

                    {/* Message text */}
                    {message.body && (
                        <p className="text-body-sm whitespace-pre-wrap break-words">
                            {message.body}
                        </p>
                    )}

                    {/* Timestamp and status */}
                    <div className={clsx(
                        "flex items-center justify-end gap-1.5 mt-1 -mb-0.5",
                        isMyMessage ? "text-white/70" : "text-text-tertiary"
                    )}>
                        {isFailed && (
                            <AlertCircle className="w-3 h-3 text-white" />
                        )}
                        <span className="text-caption">
                            {new Date(message.created_at).toLocaleTimeString([], {
                                hour: '2-digit',
                                minute: '2-digit'
                            })}
                        </span>
                        <MessageStatus status={message.status ?? 1} isMyMessage={isMyMessage} />
                    </div>
                </div>

                {/* Reactions badge */}
                {Object.keys(groupedReactions).length > 0 && (
                    <div className={clsx(
                        "flex gap-0.5 mt-1 glass rounded-full px-1.5 py-0.5 animate-scale-in"
                    )}>
                        {Object.entries(groupedReactions).map(([emoji, { count, userIds }]) => {
                            const userReacted = userIds.includes(currentUser?.id ?? -1);
                            return (
                                <button
                                    key={emoji}
                                    onClick={() => handleReactionClick(emoji)}
                                    className={clsx(
                                        "flex items-center gap-0.5 px-1 py-0.5 rounded-full transition-all duration-150",
                                        "hover:bg-bg-elevated active:scale-95",
                                        userReacted && "bg-brand-500/20"
                                    )}
                                >
                                    <span className="text-sm">{emoji}</span>
                                    {count > 1 && (
                                        <span className={clsx(
                                            "text-caption",
                                            userReacted ? "text-brand-400" : "text-text-tertiary"
                                        )}>
                                            {count}
                                        </span>
                                    )}
                                </button>
                            );
                        })}
                    </div>
                )}

                {/* Thread count */}
                {message.reply_count && message.reply_count > 0 && onViewThread && (
                    <button
                        onClick={() => onViewThread(message)}
                        className="flex items-center gap-1.5 mt-1.5 text-caption text-brand-400 hover:text-brand-300 transition-colors"
                    >
                        <MessageSquare className="w-3.5 h-3.5" />
                        {message.reply_count} {message.reply_count === 1 ? 'reply' : 'replies'}
                    </button>
                )}

                {/* Action bar */}
                <div className={clsx(
                    "absolute top-0 flex items-center gap-0.5 glass rounded-lg p-0.5 z-10 transition-all duration-150",
                    showActions ? "opacity-100 scale-100" : "opacity-0 scale-95 pointer-events-none",
                    isMyMessage ? "right-full mr-2" : "left-full ml-2"
                )}>
                    <button
                        onClick={() => setShowEmojiPicker(!showEmojiPicker)}
                        className="p-1.5 rounded-md hover:bg-bg-elevated text-text-secondary hover:text-text-primary transition-colors"
                        title="Add reaction"
                    >
                        <SmilePlus className="w-4 h-4" />
                    </button>
                    {onReply && (
                        <button
                            onClick={() => onReply(message)}
                            className="p-1.5 rounded-md hover:bg-bg-elevated text-text-secondary hover:text-text-primary transition-colors"
                            title="Reply"
                        >
                            <CornerUpLeft className="w-4 h-4" />
                        </button>
                    )}
                </div>

                {/* Emoji picker dropdown */}
                {showEmojiPicker && (
                    <div className={clsx(
                        "absolute top-8 glass rounded-xl p-2 z-20 animate-scale-in",
                        isMyMessage ? "right-full mr-2" : "left-full ml-2"
                    )}>
                        <div className="flex gap-1">
                            {QUICK_EMOJIS.map((emoji) => (
                                <button
                                    key={emoji}
                                    onClick={() => handleEmojiSelect(emoji)}
                                    className="p-1.5 hover:bg-bg-elevated rounded-lg transition-all duration-150 text-lg hover:scale-110 active:scale-95"
                                >
                                    {emoji}
                                </button>
                            ))}
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};
