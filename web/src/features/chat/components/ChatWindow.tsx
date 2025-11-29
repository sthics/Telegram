import { useState, useEffect, useRef } from 'react';
import { useChatStore } from '../stores/chatStore';
import { useAuthStore } from '@/features/auth/store';
import { chatApi } from '../api';
import { useQuery } from '@tanstack/react-query';
import { Send, Paperclip, MoreVertical, Loader2 } from 'lucide-react';
import { Button } from '@/shared/components/Button';
import { clsx } from 'clsx';

export const ChatWindow = () => {
    const activeChat = useChatStore((state) => state.activeChat);
    const currentUser = useAuthStore((state) => state.user);
    const [message, setMessage] = useState('');
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const { data: messages, isLoading } = useQuery({
        queryKey: ['messages', activeChat?.id],
        queryFn: () => chatApi.getMessages(activeChat!.id),
        enabled: !!activeChat,
        refetchInterval: 5000, // Poll every 5s until WebSocket is ready
    });

    // Scroll to bottom on new messages
    useEffect(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [messages]);

    if (!activeChat) {
        return (
            <div className="flex-1 flex items-center justify-center text-text-secondary">
                <div className="text-center">
                    <p className="bg-surface px-4 py-2 rounded-full text-sm">
                        Select a chat to start messaging
                    </p>
                </div>
            </div>
        );
    }

    return (
        <div className="flex flex-col h-full">
            {/* Header */}
            <div className="h-16 px-6 flex items-center justify-between border-b border-border-subtle bg-surface shrink-0">
                <div className="flex items-center">
                    <div className="w-10 h-10 rounded-full bg-brand-primary/20 flex items-center justify-center text-brand-primary font-medium shrink-0">
                        {activeChat.name ? activeChat.name.charAt(0) : 'U'}
                    </div>
                    <div className="ml-3">
                        <h2 className="text-base font-medium text-text-primary">
                            {activeChat.name || 'Unknown Chat'}
                        </h2>
                        <p className="text-xs text-brand-primary">Online</p>
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
                        const isMyMessage = msg.userId === currentUser?.id;
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
                                    <span
                                        className={clsx(
                                            'text-[11px] block text-right mt-1',
                                            isMyMessage ? 'text-white/70' : 'text-text-secondary'
                                        )}
                                    >
                                        {new Date(msg.createdAt).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                    </span>
                                </div>
                            </div>
                        );
                    })
                )}
                <div ref={messagesEndRef} />
            </div>

            {/* Input Area */}
            <div className="p-4 bg-surface border-t border-border-subtle shrink-0">
                <div className="flex items-end gap-2 max-w-4xl mx-auto">
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
                                    // Send message
                                    setMessage('');
                                }
                            }}
                        />
                    </div>

                    <Button
                        className="bg-brand-primary hover:bg-brand-hover text-white rounded-full w-10 h-10 p-0 flex items-center justify-center shadow-sm transition-transform active:scale-95"
                    >
                        <Send className="w-5 h-5 ml-0.5" />
                    </Button>
                </div>
            </div>
        </div>
    );
};
