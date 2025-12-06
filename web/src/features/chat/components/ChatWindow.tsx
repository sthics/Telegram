import { useState, useEffect, useRef } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Send, Paperclip, MoreVertical, Loader2, Check, CheckCheck } from 'lucide-react';
import { useChatStore } from '../stores/chatStore';
import { useAuthStore } from '@/features/auth/store';
import { chatApi } from '../api';
import type { Message } from '../types';
import { clsx } from 'clsx';
import { Button } from '@/shared/components/Button';
import { ChatInfoModal } from './ChatInfoModal';

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

    // Mark as read
    useEffect(() => {
        if (activeChat && messages && messages.length > 0) {
            const latestMsg = messages[0]; // Backend returns DESC, so first is latest
            if (latestMsg) {
                chatApi.markRead(activeChat.id, latestMsg.id);
            }
        }
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
                                Group Chat
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
                    messages?.slice().reverse().map((msg) => { // Reverse because backend returns DESC
                        const isMyMessage = msg.user_id === currentUser?.id;
                        return (
                            <div key={msg.id} className={clsx('flex', isMyMessage ? 'justify-end' : 'justify-start')}>
                                <div
                                    className={clsx(
                                        'px-4 py-2 max-w-[65%] shadow-sm',
                                        isMyMessage
                                            ? 'bg-brand-primary text-white rounded-lg rounded-tr-none'
                                            : 'bg-surface text-text-primary rounded-lg rounded-tl-none'
                                    )}
                                >
                                    <p className="text-sm whitespace-pre-wrap">{msg.body}</p>
                                    <div className={clsx(
                                        "flex items-center justify-end gap-1 mt-1",
                                        isMyMessage ? "text-white/70" : "text-text-secondary"
                                    )}>
                                        <span className="text-[11px]">
                                            {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                        </span>
                                        {isMyMessage && (
                                            msg.status === 3 ?
                                                <CheckCheck className="w-3 h-3" /> :
                                                <Check className="w-3 h-3" />
                                        )}
                                    </div>
                                </div>
                            </div>
                        );
                    })
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
