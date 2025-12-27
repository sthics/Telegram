import React from 'react';
import { clsx } from 'clsx';
import type { LucideIcon } from 'lucide-react';
import { Button } from './Button';

interface EmptyStateProps {
    icon?: LucideIcon;
    title: string;
    description?: string;
    action?: {
        label: string;
        onClick: () => void;
    };
    className?: string;
    size?: 'sm' | 'md' | 'lg';
}

export const EmptyState: React.FC<EmptyStateProps> = ({
    icon: Icon,
    title,
    description,
    action,
    className,
    size = 'md',
}) => {
    const sizes = {
        sm: {
            wrapper: 'py-8 px-4',
            icon: 'w-10 h-10',
            iconWrapper: 'w-16 h-16',
            title: 'text-body',
            description: 'text-body-sm',
        },
        md: {
            wrapper: 'py-12 px-6',
            icon: 'w-12 h-12',
            iconWrapper: 'w-20 h-20',
            title: 'text-h3',
            description: 'text-body',
        },
        lg: {
            wrapper: 'py-16 px-8',
            icon: 'w-16 h-16',
            iconWrapper: 'w-24 h-24',
            title: 'text-h2',
            description: 'text-body',
        },
    };

    const s = sizes[size];

    return (
        <div className={clsx(
            'flex flex-col items-center justify-center text-center',
            s.wrapper,
            className
        )}>
            {Icon && (
                <div className={clsx(
                    'rounded-full bg-bg-elevated flex items-center justify-center mb-4',
                    s.iconWrapper
                )}>
                    <Icon className={clsx(s.icon, 'text-text-tertiary')} />
                </div>
            )}
            <h3 className={clsx(s.title, 'font-medium text-text-primary mb-1')}>
                {title}
            </h3>
            {description && (
                <p className={clsx(s.description, 'text-text-secondary max-w-sm')}>
                    {description}
                </p>
            )}
            {action && (
                <Button
                    onClick={action.onClick}
                    className="mt-4"
                    size={size === 'sm' ? 'sm' : 'default'}
                >
                    {action.label}
                </Button>
            )}
        </div>
    );
};

EmptyState.displayName = 'EmptyState';
