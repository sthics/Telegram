import React, { useState } from 'react';
import { Eye, EyeOff, AlertCircle, Check } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    label?: string;
    error?: string;
    success?: boolean;
    hint?: string;
    leftIcon?: React.ReactNode;
    rightIcon?: React.ReactNode;
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
    ({ className, label, error, success, hint, leftIcon, rightIcon, type, ...props }, ref) => {
        const [showPassword, setShowPassword] = useState(false);
        const isPassword = type === 'password';
        const inputType = isPassword && showPassword ? 'text' : type;

        const hasStatus = error || success;
        const statusIcon = error ? (
            <AlertCircle className="w-4 h-4 text-error" />
        ) : success ? (
            <Check className="w-4 h-4 text-success" />
        ) : null;

        return (
            <div className="space-y-1.5">
                {label && (
                    <label className="block text-label font-medium text-text-secondary">
                        {label}
                    </label>
                )}
                <div className="relative">
                    {leftIcon && (
                        <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text-tertiary">
                            {leftIcon}
                        </div>
                    )}
                    <input
                        ref={ref}
                        type={inputType}
                        className={cn(
                            // Base styles
                            'flex h-10 w-full rounded-md text-body text-text-primary',
                            'bg-bg border border-border-default',
                            'placeholder:text-text-tertiary',
                            // Padding adjustments for icons
                            leftIcon ? 'pl-10' : 'px-3',
                            (rightIcon || isPassword || hasStatus) ? 'pr-10' : 'px-3',
                            'py-2',
                            // Focus states
                            'transition-all duration-150 ease-out',
                            'focus:outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20',
                            // Disabled state
                            'disabled:cursor-not-allowed disabled:opacity-50 disabled:bg-bg-raised',
                            // Error state
                            error && 'border-error focus:border-error focus:ring-error/20',
                            // Success state
                            success && 'border-success focus:border-success focus:ring-success/20',
                            className
                        )}
                        {...props}
                    />
                    {/* Right side icons */}
                    <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-2">
                        {hasStatus && statusIcon}
                        {isPassword && (
                            <button
                                type="button"
                                onClick={() => setShowPassword(!showPassword)}
                                className="text-text-tertiary hover:text-text-secondary transition-colors"
                                tabIndex={-1}
                            >
                                {showPassword ? (
                                    <EyeOff className="w-4 h-4" />
                                ) : (
                                    <Eye className="w-4 h-4" />
                                )}
                            </button>
                        )}
                        {rightIcon && !isPassword && !hasStatus && (
                            <span className="text-text-tertiary">{rightIcon}</span>
                        )}
                    </div>
                </div>
                {/* Helper text */}
                {(error || hint) && (
                    <p className={cn(
                        'text-caption',
                        error ? 'text-error' : 'text-text-tertiary'
                    )}>
                        {error || hint}
                    </p>
                )}
            </div>
        );
    }
);

Input.displayName = 'Input';
