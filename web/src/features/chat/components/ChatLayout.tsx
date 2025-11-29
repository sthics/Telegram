import React from 'react';
import { useAuthStore } from '@/features/auth/store';
import { LogOut } from 'lucide-react';
import { Button } from '@/shared/components/Button';

interface ChatLayoutProps {
    sidebar: React.ReactNode;
    children: React.ReactNode;
}

export const ChatLayout = ({ sidebar, children }: ChatLayoutProps) => {
    const logout = useAuthStore((state) => state.logout);

    return (
        <div className="flex h-screen bg-app overflow-hidden">
            {/* Sidebar */}
            <aside className="w-[320px] flex flex-col border-r border-border-subtle bg-surface">
                <div className="h-16 px-4 flex items-center justify-between border-b border-border-subtle shrink-0">
                    <h1 className="text-lg font-medium text-text-primary">Telegram</h1>
                    <Button variant="ghost" size="icon" onClick={logout} title="Logout">
                        <LogOut className="w-5 h-5" />
                    </Button>
                </div>
                <div className="flex-1 overflow-y-auto custom-scrollbar">
                    {sidebar}
                </div>
            </aside>

            {/* Main Content */}
            <main className="flex-1 flex flex-col min-w-0 bg-app relative">
                {children}
            </main>
        </div>
    );
};
