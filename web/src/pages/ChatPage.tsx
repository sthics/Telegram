import { ChatLayout } from '@/features/chat/components/ChatLayout';
import { ChatList } from '@/features/chat/components/ChatList';
import { ChatWindow } from '@/features/chat/components/ChatWindow';

export const ChatPage = () => {
    return (
        <ChatLayout sidebar={<ChatList />}>
            <ChatWindow />
        </ChatLayout>
    );
};
