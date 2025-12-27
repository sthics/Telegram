import React, { useEffect, useState } from 'react';
import { clsx } from 'clsx';

type TransitionType = 'fade' | 'slide-up' | 'slide-down' | 'slide-left' | 'slide-right' | 'scale' | 'none';

interface TransitionProps {
    show: boolean;
    children: React.ReactNode;
    type?: TransitionType;
    duration?: number;
    delay?: number;
    className?: string;
    unmountOnExit?: boolean;
    onExited?: () => void;
}

const transitions: Record<TransitionType, { enter: string; exit: string }> = {
    fade: {
        enter: 'opacity-100',
        exit: 'opacity-0',
    },
    'slide-up': {
        enter: 'opacity-100 translate-y-0',
        exit: 'opacity-0 translate-y-4',
    },
    'slide-down': {
        enter: 'opacity-100 translate-y-0',
        exit: 'opacity-0 -translate-y-4',
    },
    'slide-left': {
        enter: 'opacity-100 translate-x-0',
        exit: 'opacity-0 translate-x-4',
    },
    'slide-right': {
        enter: 'opacity-100 translate-x-0',
        exit: 'opacity-0 -translate-x-4',
    },
    scale: {
        enter: 'opacity-100 scale-100',
        exit: 'opacity-0 scale-95',
    },
    none: {
        enter: '',
        exit: '',
    },
};

export const Transition: React.FC<TransitionProps> = ({
    show,
    children,
    type = 'fade',
    duration = 150,
    delay = 0,
    className,
    unmountOnExit = true,
    onExited,
}) => {
    const [mounted, setMounted] = useState(show);
    const [visible, setVisible] = useState(show);

    useEffect(() => {
        if (show) {
            setMounted(true);
            // Small delay to allow mount before transition
            const timer = setTimeout(() => setVisible(true), 10);
            return () => clearTimeout(timer);
        } else {
            setVisible(false);
            const timer = setTimeout(() => {
                if (unmountOnExit) {
                    setMounted(false);
                }
                onExited?.();
            }, duration + delay);
            return () => clearTimeout(timer);
        }
    }, [show, duration, delay, unmountOnExit, onExited]);

    if (!mounted && unmountOnExit) return null;

    const transition = transitions[type];

    return (
        <div
            className={clsx(
                'transition-all',
                visible ? transition.enter : transition.exit,
                className
            )}
            style={{
                transitionDuration: `${duration}ms`,
                transitionDelay: `${delay}ms`,
            }}
        >
            {children}
        </div>
    );
};

Transition.displayName = 'Transition';

// Staggered children animation
interface StaggerProps {
    children: React.ReactNode;
    show?: boolean;
    staggerDelay?: number;
    className?: string;
}

export const Stagger: React.FC<StaggerProps> = ({
    children,
    show = true,
    staggerDelay = 50,
    className,
}) => {
    const childArray = React.Children.toArray(children);

    return (
        <div className={className}>
            {childArray.map((child, index) => (
                <Transition
                    key={index}
                    show={show}
                    type="slide-up"
                    delay={index * staggerDelay}
                >
                    {child}
                </Transition>
            ))}
        </div>
    );
};

Stagger.displayName = 'Stagger';
