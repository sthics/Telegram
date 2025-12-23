import { useRef, useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Send, Loader2, Paperclip, MoreVertical, X, CornerUpLeft } from 'lucide-react';
import { useAuthStore } from '@/features/auth/store';
import { useChatStore } from '../stores/chatStore';
import { chatApi } from '../api';
import type { Message } from '../types';

import { Button } from '@/shared/components/Button';
import { ChatDetailsModal } from './ChatDetailsModal';
import { MessageBubble } from './MessageBubble';

export const ChatWindow = () => {
    const activeChat = useChatStore((state) => state.activeChat);
    const currentUser = useAuthStore((state) => state.user);
    const [message, setMessage] = useState('');
    const [replyingTo, setReplyingTo] = useState<Message | null>(null);
    const messagesEndRef = useRef<HTMLDivElement>(null);
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

    // Mark as read logic
    const observer = useRef<IntersectionObserver | null>(null);
    const lastReadIdRef = useRef<number>(0);

    const handleObserver = (entries: IntersectionObserverEntry[]) => {
        if (!activeChat) return;

        entries.forEach((entry) => {
            if (entry.isIntersecting) {
                const msgId = Number(entry.target.getAttribute('data-message-id'));
                const message = messages?.find(m => m.id === msgId);

                // If it's not my message and I haven't read it yet
                if (message && message.user_id !== currentUser?.id) {
                    if (msgId > lastReadIdRef.current) {
                        lastReadIdRef.current = msgId;
                        chatApi.markRead(activeChat.id, msgId).then(() => {
                            queryClient.invalidateQueries({ queryKey: ['chats'] });
                        });
                    }
                }
            }
        });
    };

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
    }, [messages, activeChat]);

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
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages, sendMessageMutation.isPending]);

    const handleSendMessage = () => {
        if (!message.trim()) return;
        sendMessageMutation.mutate({ text: message, replyToId: replyingTo?.id });
        setMessage('');
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
            <div className="flex-1 flex items-center justify-center bg-app bg-pattern">
                <div className="text-center">
                    <p className="text-text-secondary">Select a chat to start messaging</p>
                </div>
            </div>
        );
    }

    const isGroup = activeChat.type === 2;

    return (
        <div className="flex-1 flex flex-col h-full bg-app relative">
            {/* Header */}
            <div className="h-16 px-6 flex items-center justify-between border-b border-border-subtle bg-surface shrink-0 z-10">
                <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-brand-primary text-white flex items-center justify-center text-lg font-medium">
                        {activeChat.name ? activeChat.name.charAt(0) : 'U'}
                    </div>
                    <div>
                        <h2 className="font-semibold text-text-primary">{activeChat.name}</h2>
                        {activeChat.online && (
                            <span className="text-xs text-green-500 font-medium">Online</span>
                        )}
                    </div>
                </div>

                {isGroup && (
                    <Button size="icon" variant="ghost" className="text-text-secondary hover:text-text-primary" onClick={() => setIsDetailsOpen(true)}>
                        <MoreVertical className="w-5 h-5" />
                    </Button>
                )}
            </div>

            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto p-6 space-y-4 custom-scrollbar bg-pattern">
                {isLoading ? (
                    <div className="flex justify-center pt-8">
                        <Loader2 className="w-6 h-6 animate-spin text-brand-primary" />
                    </div>
                ) : messages && messages.length > 0 ? (
                    [...messages].reverse().map((msg) => (
                        <MessageBubble
                            key={msg.id}
                            message={msg}
                            onReply={handleReply}
                            innerRef={(node) => {
                                if (node && observer.current) {
                                    observer.current.observe(node);
                                }
                            }}
                        />
                    ))
                ) : (
                    <div className="text-center pt-10">
                        <p className="text-text-secondary text-sm">No messages yet. Say hello! 👋</p>
                    </div>
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <div className="bg-surface border-t border-border-subtle">
                {/* Reply Preview Bar */}
                {replyingTo && (
                    <div className="px-4 py-2 bg-surface-hover border-b border-border-subtle flex items-center justify-between">
                        <div className="flex items-center gap-2 text-sm">
                            <CornerUpLeft className="w-4 h-4 text-brand-primary" />
                            <div className="text-text-secondary">
                                Replying to <span className="text-text-primary font-medium">message</span>
                            </div>
                            <div className="text-text-tertiary truncate max-w-[200px]">
                                {replyingTo.body}
                            </div>
                        </div>
                        <button
                            onClick={cancelReply}
                            className="p-1 hover:bg-surface rounded text-text-secondary hover:text-text-primary transition-colors"
                        >
                            <X className="w-4 h-4" />
                        </button>
                    </div>
                )}

                <div className="p-4">
                    <div className="flex items-end gap-2 max-w-4xl mx-auto">
                        <Button
                            size="icon"
                            variant="ghost"
                            className="text-text-secondary hover:text-brand-primary mb-1"
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

                        <div className="flex-1 bg-app border border-border-subtle rounded-2xl px-4 py-2 focus-within:border-brand-primary focus-within:ring-1 focus-within:ring-brand-primary transition-all">
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
                                className="w-full bg-transparent border-none focus:ring-0 resize-none max-h-32 text-text-primary placeholder:text-text-tertiary"
                                rows={1}
                                style={{ minHeight: '24px' }}
                            />
                        </div>
                        <Button
                            size="icon"
                            className="mb-1 bg-brand-primary hover:bg-brand-hover text-white shadow-sm rounded-full w-10 h-10 flex items-center justify-center transition-transform active:scale-95"
                            onClick={handleSendMessage}
                            disabled={!message.trim() && !sendMessageMutation.isPending}
                        >
                            {sendMessageMutation.isPending ? (
                                <Loader2 className="w-5 h-5 animate-spin" />
                            ) : (
                                <Send className="w-5 h-5 ml-0.5" />
                            )}
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

