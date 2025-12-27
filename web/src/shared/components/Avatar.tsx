import React, { useState } from 'react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

type AvatarSize = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl';
type AvatarStatus = 'online' | 'away' | 'busy' | 'offline';

interface AvatarProps extends React.HTMLAttributes<HTMLDivElement> {
    src?: string | null;
    name: string;
    size?: AvatarSize;
    status?: AvatarStatus;
    showStatus?: boolean;
}

const sizeClasses: Record<AvatarSize, string> = {
    xs: 'w-6 h-6 text-caption',
    sm: 'w-8 h-8 text-label',
    md: 'w-10 h-10 text-body',
    lg: 'w-12 h-12 text-h4',
    xl: 'w-16 h-16 text-h3',
    '2xl': 'w-20 h-20 text-h2',
};

const statusSizeClasses: Record<AvatarSize, string> = {
    xs: 'w-1.5 h-1.5 border',
    sm: 'w-2 h-2 border',
    md: 'w-2.5 h-2.5 border-2',
    lg: 'w-3 h-3 border-2',
    xl: 'w-3.5 h-3.5 border-2',
    '2xl': 'w-4 h-4 border-2',
};

const statusColors: Record<AvatarStatus, string> = {
    online: 'bg-success',
    away: 'bg-warning',
    busy: 'bg-error',
    offline: 'bg-text-disabled',
};

// Generate a consistent color based on name
function getAvatarColor(name: string): string {
    const colors = [
        'from-blue-500 to-blue-600',
        'from-emerald-500 to-emerald-600',
        'from-violet-500 to-violet-600',
        'from-amber-500 to-amber-600',
        'from-rose-500 to-rose-600',
        'from-cyan-500 to-cyan-600',
        'from-fuchsia-500 to-fuchsia-600',
        'from-lime-500 to-lime-600',
    ];

    let hash = 0;
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash);
    }

    return colors[Math.abs(hash) % colors.length];
}

function getInitials(name: string): string {
    const parts = name.trim().split(/\s+/);
    if (parts.length >= 2) {
        return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
}

export const Avatar: React.FC<AvatarProps> = ({
    src,
    name,
    size = 'md',
    status,
    showStatus = false,
    className,
    ...props
}) => {
    const [imageError, setImageError] = useState(false);
    const showFallback = !src || imageError;
    const initials = getInitials(name);
    const gradientColor = getAvatarColor(name);

    return (
        <div
            className={cn('relative inline-flex shrink-0', className)}
            {...props}
        >
            <div
                className={cn(
                    'rounded-full overflow-hidden flex items-center justify-center font-medium',
                    sizeClasses[size],
                    showFallback && `bg-gradient-to-br ${gradientColor} text-white`
                )}
            >
                {showFallback ? (
                    <span className="select-none">{initials}</span>
                ) : (
                    <img
                        src={src}
                        alt={name}
                        className="w-full h-full object-cover"
                        onError={() => setImageError(true)}
                    />
                )}
            </div>

            {showStatus && status && (
                <span
                    className={cn(
                        'absolute bottom-0 right-0 rounded-full border-bg-raised',
                        statusSizeClasses[size],
                        statusColors[status],
                        status === 'online' && 'animate-pulse-soft'
                    )}
                    aria-label={`Status: ${status}`}
                />
            )}
        </div>
    );
};

Avatar.displayName = 'Avatar';

// Avatar group for showing multiple avatars
interface AvatarGroupProps {
    children: React.ReactNode;
    max?: number;
    size?: AvatarSize;
    className?: string;
}

export const AvatarGroup: React.FC<AvatarGroupProps> = ({
    children,
    max = 4,
    size = 'md',
    className,
}) => {
    const avatars = React.Children.toArray(children);
    const excess = avatars.length - max;
    const visibleAvatars = avatars.slice(0, max);

    return (
        <div className={cn('flex -space-x-2', className)}>
            {visibleAvatars.map((avatar, index) => (
                <div
                    key={index}
                    className="ring-2 ring-bg-raised rounded-full"
                    style={{ zIndex: visibleAvatars.length - index }}
                >
                    {React.isValidElement(avatar)
                        ? React.cloneElement(avatar as React.ReactElement<AvatarProps>, { size })
                        : avatar
                    }
                </div>
            ))}
            {excess > 0 && (
                <div
                    className={cn(
                        'rounded-full bg-bg-overlay flex items-center justify-center font-medium text-text-secondary ring-2 ring-bg-raised',
                        sizeClasses[size]
                    )}
                >
                    +{excess}
                </div>
            )}
        </div>
    );
};

AvatarGroup.displayName = 'AvatarGroup';
