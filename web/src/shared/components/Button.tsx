import React from 'react';
import { Loader2 } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'link';
    size?: 'default' | 'sm' | 'lg' | 'icon' | 'icon-sm' | 'icon-lg';
    isLoading?: boolean;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant = 'primary', size = 'default', isLoading, children, disabled, ...props }, ref) => {
        const baseStyles = `
            inline-flex items-center justify-center
            font-medium text-body
            rounded-md
            transition-all duration-150 ease-out
            focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500/50 focus-visible:ring-offset-2 focus-visible:ring-offset-bg
            disabled:opacity-50 disabled:cursor-not-allowed disabled:pointer-events-none
            active:scale-[0.98]
        `;

        const variants = {
            primary: `
                bg-brand-500 text-white
                shadow-sm
                hover:bg-brand-600 hover:-translate-y-px hover:shadow-md
                active:bg-brand-700 active:translate-y-0 active:shadow-sm
            `,
            secondary: `
                bg-bg-raised text-text-primary
                border border-border-default
                shadow-xs
                hover:bg-bg-elevated hover:border-border-strong hover:-translate-y-px
                active:bg-bg-overlay active:translate-y-0
            `,
            danger: `
                bg-error text-white
                shadow-sm
                hover:bg-error-dark hover:-translate-y-px hover:shadow-md
                active:translate-y-0 active:shadow-sm
            `,
            ghost: `
                bg-transparent text-text-secondary
                hover:bg-bg-elevated hover:text-text-primary
                active:bg-bg-overlay
            `,
            link: `
                bg-transparent text-brand-500
                hover:text-brand-400 hover:underline
                p-0 h-auto
                active:scale-100
            `,
        };

        const sizes = {
            default: 'h-9 px-4 gap-2',
            sm: 'h-8 px-3 text-body-sm gap-1.5',
            lg: 'h-10 px-5 gap-2',
            icon: 'h-9 w-9',
            'icon-sm': 'h-8 w-8',
            'icon-lg': 'h-10 w-10',
        };

        return (
            <button
                ref={ref}
                type="button"
                disabled={disabled || isLoading}
                className={cn(baseStyles, variants[variant], sizes[size], className)}
                {...props}
            >
                {isLoading ? (
                    <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        {!size.startsWith('icon') && children}
                    </>
                ) : (
                    children
                )}
            </button>
        );
    }
);

Button.displayName = 'Button';
