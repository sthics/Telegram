import { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Send, Paperclip, MoreVertical, Loader2 } from 'lucide-react';
import { useChatStore } from '../stores/chatStore';
import { useAuthStore } from '@/features/auth/store';
import { chatApi } from '../api';
import type { Message } from '../types';

import { Button } from '@/shared/components/Button';
import { ChatInfoModal } from './ChatInfoModal';
import { MessageBubble } from './MessageBubble';

export const ChatWindow = () => {
    const activeChat = useChatStore((state) => state.activeChat);
    const currentUser = useAuthStore((state) => state.user);
    const [message, setMessage] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const [isInfoOpen, setIsInfoOpen] = useState(false);
    const queryClient = useQueryClient();

    // Fetch messages
    const { data: messages, isLoading } = useQuery({
        queryKey: ['messages', activeChat?.id],
        queryFn: () => chatApi.getMessages(activeChat!.id),
        enabled: !!activeChat,
    });

    // Fetch members for count
    const { data: members } = useQuery({
        queryKey: ['chatMembers', activeChat?.id],
        queryFn: () => chatApi.getChatMembers(activeChat!.id),
        enabled: !!activeChat && activeChat.type === 2,
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
                    // messages?.find(...) check above ensures we have the message object
                    if (msgId > lastReadIdRef.current) {
                        lastReadIdRef.current = msgId;
                        chatApi.markRead(activeChat.id, msgId).then(() => {
                            // Refresh chat list to update unread counts
                            queryClient.invalidateQueries({ queryKey: ['chats'] });
                        });
                    }
                }
            }
        });
    };

    useEffect(() => {
        // Reset lastReadId when chat changes
        lastReadIdRef.current = 0;
    }, [activeChat?.id]);

    useEffect(() => {
        if (!messages) return;

        observer.current = new IntersectionObserver(handleObserver, {
            root: null,
            rootMargin: '0px',
            threshold: 0.5,
        });

        // Observe elements
        // We need to attach ref to message elements. 
        // We'll trust the callback ref in MessageBubble
    }, [messages, activeChat]);

    const sendMessageMutation = useMutation({
        mutationFn: (text: string) => chatApi.sendMessage(activeChat!.id, text),
        onMutate: async (text) => {
            await queryClient.cancelQueries({ queryKey: ['messages', activeChat!.id] });

            const previousMessages = queryClient.getQueryData<Message[]>(['messages', activeChat!.id]);

            const optimisticMessage: Message = {
                id: Date.now(), // Temp ID
                chat_id: activeChat!.id,
                user_id: currentUser!.id,
                body: text,
                created_at: new Date().toISOString(),
                reactions: [],
                status: 1,
            };

            queryClient.setQueryData<Message[]>(['messages', activeChat!.id], (old) => {
                return [optimisticMessage, ...(old || [])];
            });

            return { previousMessages };
        },
        onError: (_err, _text, context) => {
            if (context?.previousMessages) {
                queryClient.setQueryData(['messages', activeChat!.id], context.previousMessages);
            }
        },
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ['messages', activeChat!.id] });
            queryClient.invalidateQueries({ queryKey: ['chats'] }); // Update sidebar last message
        },
    });

    // Scroll to bottom
    useEffect(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages, sendMessageMutation.isPending]); // Scroll on optimistic add too

    const handleSendMessage = () => {
        if (!message.trim()) return;
        sendMessageMutation.mutate(message);
        setMessage('');
    };

    if (!activeChat) {
        return (
            <div className="flex-1 flex items-center justify-center bg-background text-text-tertiary">
                <div className="text-center">
                    <p>Select a chat to start messaging</p>
                </div>
            </div>
        );
    }

    return (
        <div className="flex-1 flex flex-col h-full bg-background relative">
            <ChatInfoModal
                isOpen={isInfoOpen}
                onClose={() => setIsInfoOpen(false)}
                chat={activeChat}
            />

            {/* Header */}
            <div className="h-16 border-b border-border-subtle flex items-center justify-between px-6 bg-surface z-10">
                <div
                    className="flex items-center gap-3 cursor-pointer hover:opacity-80 transition-opacity"
                    onClick={() => setIsInfoOpen(true)}
                >
                    <div className="w-10 h-10 rounded-full bg-gradient-to-br from-brand-primary to-brand-hover flex items-center justify-center text-white font-medium shadow-sm">
                        {activeChat.name ? activeChat.name[0].toUpperCase() : '?'}
                    </div>
                    <div>
                        <h2 className="font-semibold text-text-primary">{activeChat.name || 'Unknown Chat'}</h2>
                        {activeChat.type === 2 && (
                            <span className="text-xs text-text-tertiary">
                                {members ? `${members.length} members` : 'Group Chat'}
                            </span>
                        )}
                        {activeChat.type === 1 && activeChat.online && (
                            <span className="text-xs text-brand-primary flex items-center gap-1">
                                <span className="w-1.5 h-1.5 rounded-full bg-brand-primary animate-pulse"></span>
                                Online
                            </span>
                        )}
                        {activeChat.type === 1 && !activeChat.online && (
                            <span className="text-xs text-text-tertiary">
                                Offline
                            </span>
                        )}
                    </div>
                </div>
                <Button variant="ghost" size="icon">
                    <MoreVertical className="w-5 h-5" />
                </Button>
            </div>

            {/* Message List */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar bg-app">
                {isLoading ? (
                    <div className="flex justify-center items-center h-full">
                        <Loader2 className="w-6 h-6 animate-spin text-brand-primary" />
                    </div>
                ) : (
                    messages?.slice().reverse().map((msg) => (
                        <MessageBubble
                            key={msg.id}
                            message={msg}
                            innerRef={(node) => {
                                if (node && observer.current) {
                                    observer.current.observe(node);
                                }
                            }}
                        />
                    ))
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <div className="p-4 bg-surface border-t border-border-subtle shrink-0">
                <div className="flex items-center gap-2 max-w-4xl mx-auto">
                    <Button variant="ghost" size="icon" className="text-text-secondary hover:text-text-primary">
                        <Paperclip className="w-5 h-5" />
                    </Button>

                    <div className="flex-1 bg-app rounded-md border border-border-subtle focus-within:border-brand-primary transition-colors">
                        <textarea
                            value={message}
                            onChange={(e) => setMessage(e.target.value)}
                            placeholder="Write a message..."
                            className="w-full bg-transparent text-sm text-text-primary placeholder:text-text-tertiary px-3 py-2.5 max-h-32 min-h-[40px] resize-none focus:outline-none custom-scrollbar"
                            rows={1}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter' && !e.shiftKey) {
                                    e.preventDefault();
                                    handleSendMessage();
                                }
                            }}
                        />
                    </div>

                    <Button
                        onClick={handleSendMessage}
                        className="bg-brand-primary hover:bg-brand-hover text-white rounded-full w-10 h-10 p-0 flex items-center justify-center shadow-sm transition-transform active:scale-95"
                    >
                        <Send className="w-5 h-5 ml-0.5" />
                    </Button>
                </div>
            </div>
        </div>
    );
};
