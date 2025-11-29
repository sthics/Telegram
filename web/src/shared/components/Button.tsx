import React from 'react';
import { Loader2 } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    size?: 'default' | 'sm' | 'lg' | 'icon';
    isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant = 'primary', size = 'default', isLoading, children, disabled, ...props }, ref) => {
        const variants = {
            primary: 'bg-brand-primary hover:bg-brand-hover text-white shadow-sm',
            secondary: 'bg-surface hover:bg-surface-hover text-text-primary border border-border-subtle',
            danger: 'bg-status-error hover:bg-red-600 text-white',
            ghost: 'bg-transparent hover:bg-surface-hover text-text-secondary hover:text-text-primary',
        };

        const sizes = {
            default: 'h-9 px-4',
            sm: 'h-8 px-3 text-xs',
            lg: 'h-10 px-8',
            icon: 'h-9 w-9',
        };

        return (
            <button
                ref={ref}
                disabled={disabled || isLoading}
                className={cn(
                    'inline-flex items-center justify-center rounded-md text-sm font-medium transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-brand-primary/50 disabled:opacity-50 disabled:cursor-not-allowed',
                    variants[variant],
                    sizes[size],
                    className
                )}
                {...props}
            >
                {isLoading && <Loader2 className="w-4 h-4 mr-2 animate-spin" />}
                {children}
            </button>
        );
    }
);
