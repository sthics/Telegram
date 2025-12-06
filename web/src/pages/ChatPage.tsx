import { ChatLayout } from '@/features/chat/components/ChatLayout';
import { ChatList } from '@/features/chat/components/ChatList';
import { ChatWindow } from '@/features/chat/components/ChatWindow';
import { WebSocketProvider } from '@/shared/providers/WebSocketProvider';

export const ChatPage = () => {
    return (
        <WebSocketProvider>
            <ChatLayout sidebar={<ChatList />}>
                <ChatWindow />
            </ChatLayout>
        </WebSocketProvider>
    );
};
