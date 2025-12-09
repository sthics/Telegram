import { clsx } from 'clsx';
import { Check, CheckCheck } from 'lucide-react';
import type { Message } from '../types';
import { useAuthStore } from '@/features/auth/store';

interface MessageBubbleProps {
    message: Message;
    // Optional ref for observer
    innerRef?: (node?: Element | null) => void;
}

export const MessageBubble = ({ message, innerRef }: MessageBubbleProps) => {
    const currentUser = useAuthStore((state) => state.user);
    const isMyMessage = message.user_id === currentUser?.id;

    return (
        <div
            ref={innerRef}
            className={clsx('flex', isMyMessage ? 'justify-end' : 'justify-start')}
            data-message-id={message.id}
        >
            <div
                className={clsx(
                    'px-4 py-2 max-w-[65%] shadow-sm',
                    isMyMessage
                        ? 'bg-brand-primary text-white rounded-lg rounded-tr-none'
                        : 'bg-surface text-text-primary rounded-lg rounded-tl-none'
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
        </div>
    );
};
