import React from 'react';
import { clsx } from 'clsx';

interface SpinnerProps {
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    className?: string;
}

const sizes = {
    xs: 'w-3 h-3',
    sm: 'w-4 h-4',
    md: 'w-6 h-6',
    lg: 'w-8 h-8',
    xl: 'w-12 h-12',
};

export const Spinner: React.FC<SpinnerProps> = ({ size = 'md', className }) => (
    <svg
        className={clsx('animate-spin', sizes[size], className)}
        xmlns="http://www.w3.org/2000/svg"
        fill="none"
        viewBox="0 0 24 24"
        aria-label="Loading"
    >
        <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
        />
        <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
        />
    </svg>
);

Spinner.displayName = 'Spinner';

// Dots spinner alternative
export const DotsSpinner: React.FC<{ className?: string }> = ({ className }) => (
    <div className={clsx('flex items-center justify-center gap-1', className)}>
        <div className="w-2 h-2 rounded-full bg-current animate-bounce" style={{ animationDelay: '0ms' }} />
        <div className="w-2 h-2 rounded-full bg-current animate-bounce" style={{ animationDelay: '150ms' }} />
        <div className="w-2 h-2 rounded-full bg-current animate-bounce" style={{ animationDelay: '300ms' }} />
    </div>
);

DotsSpinner.displayName = 'DotsSpinner';

// Pulse spinner
export const PulseSpinner: React.FC<{ size?: 'sm' | 'md' | 'lg'; className?: string }> = ({
    size = 'md',
    className
}) => {
    const sizeClasses = {
        sm: 'w-8 h-8',
        md: 'w-12 h-12',
        lg: 'w-16 h-16',
    };

    return (
        <div className={clsx('relative', sizeClasses[size], className)}>
            <div className="absolute inset-0 rounded-full bg-brand-500/30 animate-ping" />
            <div className="absolute inset-2 rounded-full bg-brand-500/50 animate-pulse" />
            <div className="absolute inset-4 rounded-full bg-brand-500" />
        </div>
    );
};

PulseSpinner.displayName = 'PulseSpinner';
