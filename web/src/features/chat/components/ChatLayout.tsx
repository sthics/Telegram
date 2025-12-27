import React, { useState } from 'react';
import { useAuthStore } from '@/features/auth/store';
import { useChatStore } from '../stores/chatStore';
import { LogOut, ArrowLeft } from 'lucide-react';
import { Button } from '@/shared/components/Button';
import { clsx } from 'clsx';

interface ChatLayoutProps {
    sidebar: React.ReactNode;
    children: React.ReactNode;
}

export const ChatLayout = ({ sidebar, children }: ChatLayoutProps) => {
    const logout = useAuthStore((state) => state.logout);
    const { activeChat, setActiveChat } = useChatStore();
    const [sidebarOpen, setSidebarOpen] = useState(false);

    // On mobile, show sidebar when no chat is selected
    const showMobileSidebar = !activeChat;

    const handleBackToList = () => {
        setActiveChat(null);
    };

    return (
        <div className="flex h-screen bg-bg overflow-hidden">
            {/* Mobile overlay */}
            {sidebarOpen && (
                <div
                    className="fixed inset-0 bg-bg-base/60 backdrop-blur-sm z-overlay lg:hidden"
                    onClick={() => setSidebarOpen(false)}
                />
            )}

            {/* Sidebar */}
            <aside
                className={clsx(
                    "flex flex-col border-r border-border-subtle bg-bg-raised z-overlay",
                    // Mobile: full width overlay or hidden
                    "fixed inset-y-0 left-0 w-full max-w-[320px]",
                    "lg:relative lg:w-[320px] lg:max-w-none",
                    // Mobile visibility
                    "transition-transform duration-200 ease-out lg:translate-x-0",
                    showMobileSidebar || sidebarOpen
                        ? "translate-x-0"
                        : "-translate-x-full lg:translate-x-0"
                )}
            >
                {/* Header */}
                <div className="h-14 px-4 flex items-center justify-between border-b border-border-subtle shrink-0">
                    <h1 className="text-h3 text-text-primary font-semibold">Telegram</h1>
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={logout}
                        title="Logout"
                        className="rounded-full"
                    >
                        <LogOut className="w-5 h-5" />
                    </Button>
                </div>

                {/* Sidebar content */}
                <div className="flex-1 overflow-y-auto">
                    {sidebar}
                </div>
            </aside>

            {/* Main Content */}
            <main className={clsx(
                "flex-1 flex flex-col min-w-0 bg-bg relative",
                // Hide on mobile when sidebar is shown
                showMobileSidebar && "hidden lg:flex"
            )}>
                {/* Mobile header with back button */}
                {activeChat && (
                    <div className="lg:hidden absolute top-0 left-0 z-20 p-2">
                        <Button
                            variant="ghost"
                            size="icon-sm"
                            onClick={handleBackToList}
                            className="rounded-full bg-bg-raised/80 backdrop-blur-sm"
                        >
                            <ArrowLeft className="w-5 h-5" />
                        </Button>
                    </div>
                )}
                {children}
            </main>
        </div>
    );
};
