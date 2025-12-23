import { useState, useRef, useEffect } from 'react';
import { SmilePlus } from 'lucide-react';

interface EmojiPickerProps {
    onSelect: (emoji: string) => void;
    disabled?: boolean;
}

const QUICK_EMOJIS = ['👍', '❤️', '😂', '😮', '😢', '👎', '🎉', '🔥'];

export const EmojiPicker = ({ onSelect, disabled }: EmojiPickerProps) => {
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    // Close picker on outside click
    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setIsOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleSelect = (emoji: string) => {
        onSelect(emoji);
        setIsOpen(false);
    };

    return (
        <div ref={containerRef} className="relative">
            <button
                onClick={() => setIsOpen(!isOpen)}
                disabled={disabled}
                className="p-1 rounded hover:bg-surface-hover text-text-secondary hover:text-text-primary transition-colors disabled:opacity-50"
                title="Add reaction"
            >
                <SmilePlus className="w-4 h-4" />
            </button>

            {isOpen && (
                <div className="absolute bottom-full right-0 mb-1 bg-surface border border-border-subtle rounded-lg shadow-lg p-2 z-50">
                    <div className="flex gap-1">
                        {QUICK_EMOJIS.map((emoji) => (
                            <button
                                key={emoji}
                                onClick={() => handleSelect(emoji)}
                                className="p-1.5 hover:bg-surface-hover rounded transition-colors text-lg"
                            >
                                {emoji}
                            </button>
                        ))}
                    </div>
                </div>
            )}
        </div>
    );
};
