import React from 'react';
import { clsx } from 'clsx';

interface TypingIndicatorProps {
    className?: string;
    names?: string[];
}

export const TypingIndicator: React.FC<TypingIndicatorProps> = ({
    className,
    names = [],
}) => {
    const getTypingText = () => {
        if (names.length === 0) return 'typing';
        if (names.length === 1) return `${names[0]} is typing`;
        if (names.length === 2) return `${names[0]} and ${names[1]} are typing`;
        return `${names[0]} and ${names.length - 1} others are typing`;
    };

    return (
        <div className={clsx('flex items-center gap-2', className)}>
            <div className="flex items-center gap-0.5">
                <span className="typing-dot" />
                <span className="typing-dot" />
                <span className="typing-dot" />
            </div>
            <span className="text-caption text-text-tertiary">
                {getTypingText()}
            </span>
        </div>
    );
};

TypingIndicator.displayName = 'TypingIndicator';

// Inline typing dots (without text)
export const TypingDots: React.FC<{ className?: string }> = ({ className }) => (
    <div className={clsx('flex items-center gap-1 px-3 py-2', className)}>
        <span className="w-2 h-2 rounded-full bg-text-tertiary animate-typing-1" />
        <span className="w-2 h-2 rounded-full bg-text-tertiary animate-typing-2" />
        <span className="w-2 h-2 rounded-full bg-text-tertiary animate-typing-3" />
    </div>
);

TypingDots.displayName = 'TypingDots';
