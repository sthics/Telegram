import { useRef, useState, useEffect, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Send, Paperclip, MoreVertical, X, ChevronDown, MessageCircle } from 'lucide-react';
import { useAuthStore } from '@/features/auth/store';
import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import type { Message } from '../types';
import { clsx } from 'clsx';

import { Button } from '@/shared/components/Button';
import { Avatar } from '@/shared/components/Avatar';
import { SkeletonMessage } from '@/shared/components/Skeleton';
import { ChatDetailsModal } from './ChatDetailsModal';
import { MessageBubble } from './MessageBubble';

// Format date for date separators
const formatDateSeparator = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = now.getTime() - date.getTime();
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) {
        return date.toLocaleDateString([], { weekday: 'long' });
    }
    return date.toLocaleDateString([], {
        month: 'long',
        day: 'numeric',
        year: now.getFullYear() !== date.getFullYear() ? 'numeric' : undefined
    });
};

// Check if two dates are on the same day
const isSameDay = (date1: string, date2: string) => {
    const d1 = new Date(date1);
    const d2 = new Date(date2);
    return d1.toDateString() === d2.toDateString();
};

// Group messages by sender and time proximity
interface GroupedMessage {
    message: Message;
    isFirstInGroup: boolean;
    isLastInGroup: boolean;
    showDateSeparator: boolean;
}

const groupMessages = (messages: Message[]): GroupedMessage[] => {
    if (!messages || messages.length === 0) return [];

    const reversed = [...messages].reverse();
    const grouped: GroupedMessage[] = [];

    reversed.forEach((message, index) => {
        const prevMessage = index > 0 ? reversed[index - 1] : null;
        const nextMessage = index < reversed.length - 1 ? reversed[index + 1] : null;

        // Check if this is a new day
        const showDateSeparator = !prevMessage || !isSameDay(prevMessage.created_at, message.created_at);

        // Check if this is the first/last message in a group from the same sender
        const timeDiff = prevMessage
            ? new Date(message.created_at).getTime() - new Date(prevMessage.created_at).getTime()
            : Infinity;
        const nextTimeDiff = nextMessage
            ? new Date(nextMessage.created_at).getTime() - new Date(message.created_at).getTime()
            : Infinity;

        const isFirstInGroup = !prevMessage ||
            prevMessage.user_id !== message.user_id ||
            timeDiff > 60000 || // More than 1 minute apart
            showDateSeparator;

        const isLastInGroup = !nextMessage ||
            nextMessage.user_id !== message.user_id ||
            nextTimeDiff > 60000 ||
            (nextMessage && !isSameDay(message.created_at, nextMessage.created_at));

        grouped.push({
            message,
            isFirstInGroup,
            isLastInGroup,
            showDateSeparator,
        });
    });

    return grouped;
};

export const ChatWindow = () => {
    const activeChat = useChatStore((state) => state.activeChat);
    const currentUser = useAuthStore((state) => state.user);
    const [message, setMessage] = useState('');
    const [replyingTo, setReplyingTo] = useState<Message | null>(null);
    const [showScrollButton, setShowScrollButton] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const messagesContainerRef = useRef<HTMLDivElement>(null);
    const [isDetailsOpen, setIsDetailsOpen] = useState(false);
    const queryClient = useQueryClient();
    const fileInputRef = useRef<HTMLInputElement>(null);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    // Fetch messages
    const { data: messages, isLoading } = useQuery({
        queryKey: ['messages', activeChat?.id],
        queryFn: () => chatApi.getMessages(activeChat!.id),
        enabled: !!activeChat,
    });

    // Group messages
    const groupedMessages = useMemo(() => groupMessages(messages || []), [messages]);

    // Mark as read logic
    const observer = useRef<IntersectionObserver | null>(null);
    const lastReadIdRef = useRef<number>(0);

    const handleObserver = useCallback((entries: IntersectionObserverEntry[]) => {
        if (!activeChat || !messages) return;

        entries.forEach((entry) => {
            if (entry.isIntersecting) {
                const msgId = Number(entry.target.getAttribute('data-message-id'));
                const msg = messages.find(m => m.id === msgId);

                if (msg && msg.user_id !== currentUser?.id) {
                    if (msgId > lastReadIdRef.current) {
                        lastReadIdRef.current = msgId;
                        chatApi.markRead(activeChat.id, msgId).then(() => {
                            queryClient.invalidateQueries({ queryKey: ['chats'] });
                        });
                    }
                }
            }
        });
    }, [activeChat, messages, currentUser?.id, queryClient]);

    useEffect(() => {
        lastReadIdRef.current = 0;
        setReplyingTo(null);
    }, [activeChat?.id]);

    useEffect(() => {
        if (!messages) return;
        observer.current = new IntersectionObserver(handleObserver, {
            root: null,
            rootMargin: '0px',
            threshold: 0.5,
        });
        return () => observer.current?.disconnect();
    }, [messages, handleObserver]);

    // Handle scroll position for "new messages" button
    const handleScroll = useCallback(() => {
        if (!messagesContainerRef.current) return;
        const { scrollTop, scrollHeight, clientHeight } = messagesContainerRef.current;
        const isNearBottom = scrollHeight - scrollTop - clientHeight < 200;
        setShowScrollButton(!isNearBottom);
    }, []);

    const scrollToBottom = useCallback((smooth = true) => {
        messagesEndRef.current?.scrollIntoView({
            behavior: smooth ? 'smooth' : 'auto'
        });
    }, []);

    // Send Message Mutation
    const sendMessageMutation = useMutation({
        mutationFn: ({ text, mediaUrl, replyToId }: { text: string; mediaUrl?: string; replyToId?: number }) =>
            replyToId
                ? chatApi.sendReply(activeChat!.id, replyToId, text, mediaUrl)
                : chatApi.sendMessage(activeChat!.id, text, mediaUrl),
        onMutate: async ({ text, mediaUrl, replyToId }) => {
            await queryClient.cancelQueries({ queryKey: ['messages', activeChat!.id] });
            const previousMessages = queryClient.getQueryData<Message[]>(['messages', activeChat!.id]);
            const optimisticMessage: Message = {
                id: Date.now(),
                chat_id: activeChat!.id,
                user_id: currentUser!.id,
                body: text,
                media_url: mediaUrl,
                reply_to_id: replyToId,
                created_at: new Date().toISOString(),
                reactions: [],
                status: 1,
            };
            queryClient.setQueryData<Message[]>(['messages', activeChat!.id], (old) => [optimisticMessage, ...(old || [])]);
            return { previousMessages };
        },
        onError: (_err, _vars, context) => {
            if (context?.previousMessages) {
                queryClient.setQueryData(['messages', activeChat!.id], context.previousMessages);
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['messages', activeChat!.id] });
            queryClient.invalidateQueries({ queryKey: ['chats'] });
            setReplyingTo(null);
        },
    });

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        try {
            const { uploadUrl, objectKey } = await chatApi.getPresignedUrl(file.name, file.type || 'application/octet-stream');
            await chatApi.uploadFileToUrl(uploadUrl, file, file.type || 'application/octet-stream');
            const publicUrl = `http://localhost:9000/chat-media/${objectKey}`;
            sendMessageMutation.mutate({ text: file.name, mediaUrl: publicUrl, replyToId: replyingTo?.id });
        } catch (error) {
            console.error('Failed to upload file:', error);
        } finally {
            if (fileInputRef.current) fileInputRef.current.value = '';
        }
    };

    useEffect(() => {
        if (!showScrollButton) {
            scrollToBottom(sendMessageMutation.isPending);
        }
    }, [messages, sendMessageMutation.isPending, showScrollButton, scrollToBottom]);

    // Auto-resize textarea
    useEffect(() => {
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
            textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 160)}px`;
        }
    }, [message]);

    const handleSendMessage = () => {
        if (!message.trim()) return;
        sendMessageMutation.mutate({ text: message, replyToId: replyingTo?.id });
        setMessage('');
        if (textareaRef.current) {
            textareaRef.current.style.height = 'auto';
        }
    };

    const handleReply = (msg: Message) => {
        setReplyingTo(msg);
        textareaRef.current?.focus();
    };

    const cancelReply = () => {
        setReplyingTo(null);
    };

    if (!activeChat) {
        return (
            <div className="flex-1 flex items-center justify-center bg-bg">
                <div className="empty-state">
                    <MessageCircle className="w-16 h-16 mb-4 text-text-tertiary opacity-50" />
                    <h2 className="text-h2 text-text-primary mb-2">Select a conversation</h2>
                    <p className="text-body text-text-secondary">
                        Choose a chat from the sidebar to start messaging
                    </p>
                </div>
            </div>
        );
    }

    const isGroup = activeChat.type === 2;

    return (
        <div className="flex-1 flex flex-col h-full bg-bg relative">
            {/* Header */}
            <header className="h-14 px-4 flex items-center justify-between border-b border-border-subtle bg-bg-raised shrink-0 z-10">
                <div className="flex items-center gap-3">
                    <Avatar
                        name={activeChat.name || 'Unknown'}
                        size="md"
                        status={activeChat.online ? 'online' : undefined}
                        showStatus={!isGroup}
                    />
                    <div>
                        <h2 className="text-h4 text-text-primary">{activeChat.name}</h2>
                        <p className="text-caption text-text-tertiary">
                            {activeChat.online ? (
                                <span className="text-success">Online</span>
                            ) : isGroup ? (
                                'Group chat'
                            ) : (
                                'Offline'
                            )}
                        </p>
                    </div>
                </div>

                {isGroup && (
                    <Button
                        size="icon-sm"
                        variant="ghost"
                        onClick={() => setIsDetailsOpen(true)}
                        className="rounded-full"
                    >
                        <MoreVertical className="w-5 h-5" />
                    </Button>
                )}
            </header>

            {/* Messages Area */}
            <div
                ref={messagesContainerRef}
                onScroll={handleScroll}
                className="flex-1 overflow-y-auto px-4 py-4"
            >
                {isLoading ? (
                    <div className="space-y-4">
                        {Array.from({ length: 6 }).map((_, i) => (
                            <SkeletonMessage key={i} isOwn={i % 3 === 0} />
                        ))}
                    </div>
                ) : groupedMessages.length > 0 ? (
                    <div className="max-w-3xl mx-auto">
                        {groupedMessages.map(({ message: msg, isFirstInGroup, isLastInGroup, showDateSeparator }) => (
                            <div key={msg.id}>
                                {/* Date separator */}
                                {showDateSeparator && (
                                    <div className="flex items-center justify-center my-4">
                                        <div className="glass px-3 py-1 rounded-full">
                                            <span className="text-caption text-text-secondary">
                                                {formatDateSeparator(msg.created_at)}
                                            </span>
                                        </div>
                                    </div>
                                )}
                                <MessageBubble
                                    message={msg}
                                    onReply={handleReply}
                                    isFirstInGroup={isFirstInGroup}
                                    isLastInGroup={isLastInGroup}
                                    showAvatar={true}
                                    innerRef={(node) => {
                                        if (node && observer.current) {
                                            observer.current.observe(node);
                                        }
                                    }}
                                />
                            </div>
                        ))}
                    </div>
                ) : (
                    <div className="empty-state h-full">
                        <div className="w-16 h-16 rounded-full bg-bg-elevated flex items-center justify-center mb-4">
                            <MessageCircle className="w-8 h-8 text-text-tertiary" />
                        </div>
                        <p className="text-body text-text-secondary">No messages yet</p>
                        <p className="text-body-sm text-text-tertiary mt-1">Send a message to start the conversation</p>
                    </div>
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Scroll to bottom button */}
            {showScrollButton && (
                <button
                    onClick={() => scrollToBottom()}
                    className="absolute bottom-24 right-6 glass rounded-full p-2 shadow-lg animate-fade-in hover:bg-bg-elevated transition-colors"
                    aria-label="Scroll to bottom"
                >
                    <ChevronDown className="w-5 h-5 text-text-secondary" />
                </button>
            )}

            {/* Input Area */}
            <div className="bg-bg-raised border-t border-border-subtle">
                {/* Reply Preview Bar */}
                {replyingTo && (
                    <div className="px-4 py-2 bg-bg-elevated border-b border-border-subtle flex items-center justify-between animate-slide-down">
                        <div className="flex items-center gap-3 text-body-sm min-w-0">
                            <div className="w-1 h-8 bg-brand-500 rounded-full shrink-0" />
                            <div className="min-w-0">
                                <p className="text-text-secondary">
                                    Replying to <span className="text-text-primary font-medium">message</span>
                                </p>
                                <p className="text-text-tertiary truncate">
                                    {replyingTo.body}
                                </p>
                            </div>
                        </div>
                        <Button
                            size="icon-sm"
                            variant="ghost"
                            onClick={cancelReply}
                            className="rounded-full shrink-0"
                        >
                            <X className="w-4 h-4" />
                        </Button>
                    </div>
                )}

                <div className="p-3">
                    <div className="flex items-end gap-2 max-w-3xl mx-auto">
                        <Button
                            size="icon"
                            variant="ghost"
                            className="shrink-0 rounded-full"
                            onClick={() => fileInputRef.current?.click()}
                            disabled={sendMessageMutation.isPending}
                        >
                            <Paperclip className="w-5 h-5" />
                        </Button>
                        <input
                            type="file"
                            ref={fileInputRef}
                            className="hidden"
                            onChange={handleFileSelect}
                        />

                        <div className="flex-1 bg-bg border border-border-default rounded-2xl px-4 py-2.5 focus-within:border-brand-500 focus-within:ring-2 focus-within:ring-brand-500/20 transition-all">
                            <textarea
                                ref={textareaRef}
                                value={message}
                                onChange={(e) => setMessage(e.target.value)}
                                onKeyDown={(e) => {
                                    if (e.key === 'Enter' && !e.shiftKey) {
                                        e.preventDefault();
                                        handleSendMessage();
                                    }
                                    if (e.key === 'Escape' && replyingTo) {
                                        cancelReply();
                                    }
                                }}
                                placeholder={replyingTo ? "Type your reply..." : "Type a message..."}
                                className="w-full bg-transparent border-none focus:ring-0 focus:outline-none resize-none text-body text-text-primary placeholder:text-text-tertiary"
                                rows={1}
                                style={{ minHeight: '24px', maxHeight: '160px' }}
                            />
                        </div>

                        <Button
                            size="icon"
                            className={clsx(
                                "shrink-0 rounded-full transition-all duration-200",
                                message.trim()
                                    ? "bg-brand-500 hover:bg-brand-600 text-white shadow-sm"
                                    : "bg-bg-elevated text-text-tertiary"
                            )}
                            onClick={handleSendMessage}
                            disabled={!message.trim() || sendMessageMutation.isPending}
                            isLoading={sendMessageMutation.isPending}
                        >
                            <Send className="w-5 h-5" />
                        </Button>
                    </div>
                </div>
            </div>

            {activeChat && (
                <ChatDetailsModal
                    isOpen={isDetailsOpen}
                    onClose={() => setIsDetailsOpen(false)}
                    chat={activeChat}
                />
            )}
        </div>
    );
};
