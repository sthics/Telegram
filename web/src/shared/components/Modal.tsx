import React, { useEffect, useState, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { X } from 'lucide-react';
import { Button } from './Button';

interface ModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
    size?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
}

export const Modal: React.FC<ModalProps> = ({
    isOpen,
    onClose,
    title,
    children,
    size = 'md'
}) => {
    const [mounted, setMounted] = useState(false);
    const [isAnimating, setIsAnimating] = useState(false);

    const sizeClasses = {
        sm: 'max-w-sm',
        md: 'max-w-md',
        lg: 'max-w-lg',
        xl: 'max-w-xl',
        full: 'max-w-[calc(100vw-2rem)] max-h-[calc(100vh-2rem)]',
    };

    // Handle escape key
    const handleEscape = useCallback((e: KeyboardEvent) => {
        if (e.key === 'Escape') {
            onClose();
        }
    }, [onClose]);

    useEffect(() => {
        setMounted(true);
    }, []);

    useEffect(() => {
        if (isOpen) {
            setIsAnimating(true);
            document.body.style.overflow = 'hidden';
            document.addEventListener('keydown', handleEscape);
        } else {
            document.body.style.overflow = 'unset';
        }

        return () => {
            document.body.style.overflow = 'unset';
            document.removeEventListener('keydown', handleEscape);
        };
    }, [isOpen, handleEscape]);

    if (!mounted) return null;
    if (!isOpen && !isAnimating) return null;

    const handleAnimationEnd = () => {
        if (!isOpen) {
            setIsAnimating(false);
        }
    };

    return createPortal(
        <div
            className="fixed inset-0 z-modal flex items-center justify-center p-4"
            role="dialog"
            aria-modal="true"
            aria-labelledby="modal-title"
        >
            {/* Backdrop */}
            <div
                className={`
                    absolute inset-0 bg-bg-base/80 backdrop-blur-sm
                    transition-opacity duration-200
                    ${isOpen ? 'opacity-100' : 'opacity-0'}
                `}
                onClick={onClose}
                aria-hidden="true"
            />

            {/* Dialog */}
            <div
                className={`
                    relative bg-bg-raised rounded-xl shadow-xl w-full
                    border border-border-subtle overflow-hidden
                    transition-all duration-200 ease-out
                    ${isOpen
                        ? 'opacity-100 scale-100 translate-y-0'
                        : 'opacity-0 scale-95 translate-y-2'
                    }
                    ${sizeClasses[size]}
                `}
                onAnimationEnd={handleAnimationEnd}
            >
                {/* Header */}
                <div className="flex items-center justify-between px-6 py-4 border-b border-border-subtle">
                    <h3
                        id="modal-title"
                        className="text-h3 text-text-primary"
                    >
                        {title}
                    </h3>
                    <Button
                        variant="ghost"
                        size="icon-sm"
                        onClick={onClose}
                        className="rounded-full -mr-2"
                        aria-label="Close modal"
                    >
                        <X className="w-5 h-5 text-text-secondary" />
                    </Button>
                </div>

                {/* Content */}
                <div className="p-6 overflow-y-auto max-h-[calc(100vh-12rem)]">
                    {children}
                </div>
            </div>
        </div>,
        document.body
    );
};

Modal.displayName = 'Modal';
