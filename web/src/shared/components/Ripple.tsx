import React, { useState, useCallback } from 'react';
import { clsx } from 'clsx';

interface RippleProps {
    children: React.ReactNode;
    className?: string;
    color?: string;
    disabled?: boolean;
}

interface RippleEffect {
    id: number;
    x: number;
    y: number;
    size: number;
}

let rippleId = 0;

export const Ripple: React.FC<RippleProps> = ({
    children,
    className,
    color = 'rgba(255, 255, 255, 0.3)',
    disabled = false,
}) => {
    const [ripples, setRipples] = useState<RippleEffect[]>([]);

    const handleClick = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
        if (disabled) return;

        const rect = e.currentTarget.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        const size = Math.max(rect.width, rect.height) * 2;

        const newRipple: RippleEffect = {
            id: ++rippleId,
            x: x - size / 2,
            y: y - size / 2,
            size,
        };

        setRipples((prev) => [...prev, newRipple]);

        // Remove ripple after animation
        setTimeout(() => {
            setRipples((prev) => prev.filter((r) => r.id !== newRipple.id));
        }, 600);
    }, [disabled]);

    return (
        <div
            className={clsx('relative overflow-hidden', className)}
            onClick={handleClick}
        >
            {children}
            {ripples.map((ripple) => (
                <span
                    key={ripple.id}
                    className="absolute rounded-full pointer-events-none animate-ripple"
                    style={{
                        left: ripple.x,
                        top: ripple.y,
                        width: ripple.size,
                        height: ripple.size,
                        backgroundColor: color,
                    }}
                />
            ))}
        </div>
    );
};

Ripple.displayName = 'Ripple';
