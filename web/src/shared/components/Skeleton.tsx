import React from 'react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
    variant?: 'text' | 'circular' | 'rectangular';
    width?: string | number;
    height?: string | number;
    lines?: number;
    animate?: boolean;
}

export const Skeleton: React.FC<SkeletonProps> = ({
    className,
    variant = 'rectangular',
    width,
    height,
    lines = 1,
    animate = true,
    style,
    ...props
}) => {
    const baseStyles = cn(
        'bg-bg-elevated rounded',
        animate && 'skeleton',
        className
    );

    const computedStyle = {
        width: typeof width === 'number' ? `${width}px` : width,
        height: typeof height === 'number' ? `${height}px` : height,
        ...style,
    };

    if (variant === 'circular') {
        return (
            <div
                className={cn(baseStyles, 'rounded-full')}
                style={{
                    ...computedStyle,
                    width: computedStyle.width || computedStyle.height || '40px',
                    height: computedStyle.height || computedStyle.width || '40px',
                }}
                {...props}
            />
        );
    }

    if (variant === 'text') {
        if (lines > 1) {
            return (
                <div className="space-y-2" {...props}>
                    {Array.from({ length: lines }).map((_, i) => (
                        <div
                            key={i}
                            className={cn(baseStyles, 'h-4')}
                            style={{
                                ...computedStyle,
                                width: i === lines - 1 ? '75%' : computedStyle.width || '100%',
                                animationDelay: `${i * 100}ms`,
                            }}
                        />
                    ))}
                </div>
            );
        }

        return (
            <div
                className={cn(baseStyles, 'h-4')}
                style={{
                    ...computedStyle,
                    width: computedStyle.width || '100%',
                }}
                {...props}
            />
        );
    }

    // Default: rectangular
    return (
        <div
            className={baseStyles}
            style={{
                ...computedStyle,
                width: computedStyle.width || '100%',
                height: computedStyle.height || '20px',
            }}
            {...props}
        />
    );
};

Skeleton.displayName = 'Skeleton';

// Preset skeleton components for common use cases
export const SkeletonAvatar: React.FC<{
    size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
    className?: string;
}> = ({ size = 'md', className }) => {
    const sizes = {
        xs: 24,
        sm: 32,
        md: 40,
        lg: 48,
        xl: 64,
    };

    return (
        <Skeleton
            variant="circular"
            width={sizes[size]}
            height={sizes[size]}
            className={className}
        />
    );
};

SkeletonAvatar.displayName = 'SkeletonAvatar';

export const SkeletonText: React.FC<{
    lines?: number;
    width?: string | number;
    className?: string;
}> = ({ lines = 1, width, className }) => (
    <Skeleton
        variant="text"
        lines={lines}
        width={width}
        className={className}
    />
);

SkeletonText.displayName = 'SkeletonText';

export const SkeletonButton: React.FC<{
    size?: 'sm' | 'default' | 'lg';
    className?: string;
}> = ({ size = 'default', className }) => {
    const sizes = {
        sm: { width: 64, height: 32 },
        default: { width: 80, height: 36 },
        lg: { width: 96, height: 40 },
    };

    return (
        <Skeleton
            width={sizes[size].width}
            height={sizes[size].height}
            className={cn('rounded-md', className)}
        />
    );
};

SkeletonButton.displayName = 'SkeletonButton';

// Chat list item skeleton
export const SkeletonChatItem: React.FC<{ className?: string }> = ({ className }) => (
    <div className={cn('flex items-center gap-3 px-4 py-3', className)}>
        <SkeletonAvatar size="lg" />
        <div className="flex-1 space-y-2">
            <Skeleton variant="text" width="60%" />
            <Skeleton variant="text" width="80%" />
        </div>
        <Skeleton variant="text" width={40} />
    </div>
);

SkeletonChatItem.displayName = 'SkeletonChatItem';

// Message skeleton
export const SkeletonMessage: React.FC<{
    isOwn?: boolean;
    className?: string;
}> = ({ isOwn = false, className }) => (
    <div
        className={cn(
            'flex gap-2 max-w-[70%]',
            isOwn ? 'ml-auto flex-row-reverse' : '',
            className
        )}
    >
        {!isOwn && <SkeletonAvatar size="sm" />}
        <div className="space-y-1">
            <Skeleton
                width={Math.floor(Math.random() * 100) + 100}
                height={36}
                className="rounded-2xl"
            />
        </div>
    </div>
);

SkeletonMessage.displayName = 'SkeletonMessage';
