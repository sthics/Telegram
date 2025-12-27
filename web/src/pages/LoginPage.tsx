import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { MessageCircle, AlertCircle } from 'lucide-react';
import { useAuthStore } from '@/features/auth/store';
import { authApi } from '@/features/auth/api';
import { Button } from '@/shared/components/Button';
import { Input } from '@/shared/components/Input';
import { Card, CardContent, CardHeader } from '@/shared/components/Card';

export const LoginPage = () => {
    const navigate = useNavigate();
    const setAuth = useAuthStore((state) => state.setAuth);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');
    const [shake, setShake] = useState(false);

    const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault();
        setIsLoading(true);
        setError('');

        const formData = new FormData(e.currentTarget);
        const email = formData.get('email') as string;
        const password = formData.get('password') as string;

        try {
            const { accessToken, user } = await authApi.login({ email, password });
            setAuth(accessToken, user);
            navigate('/');
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to login');
            setShake(true);
            setTimeout(() => setShake(false), 500);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-bg p-4">
            {/* Background decoration */}
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                <div className="absolute -top-40 -right-40 w-80 h-80 bg-brand-500/10 rounded-full blur-3xl" />
                <div className="absolute -bottom-40 -left-40 w-80 h-80 bg-brand-600/10 rounded-full blur-3xl" />
            </div>

            <Card
                className={`w-full max-w-md relative animate-fade-in ${shake ? 'animate-shake' : ''}`}
                variant="elevated"
            >
                <CardHeader className="text-center pb-2">
                    {/* Logo */}
                    <div className="mx-auto w-16 h-16 bg-gradient-to-br from-brand-400 to-brand-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg shadow-brand-500/25">
                        <MessageCircle className="w-8 h-8 text-white" />
                    </div>
                    <h1 className="text-h1 text-text-primary">Welcome back</h1>
                    <p className="text-body text-text-secondary mt-1">
                        Sign in to continue to Telegram
                    </p>
                </CardHeader>

                <CardContent className="pt-4">
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {/* Error message */}
                        {error && (
                            <div className="flex items-center gap-2 p-3 rounded-lg bg-error/10 border border-error/20 text-error animate-slide-down">
                                <AlertCircle className="w-4 h-4 shrink-0" />
                                <p className="text-body-sm">{error}</p>
                            </div>
                        )}

                        <Input
                            name="email"
                            type="email"
                            label="Email"
                            placeholder="Enter your email"
                            required
                            autoComplete="email"
                        />

                        <Input
                            name="password"
                            type="password"
                            label="Password"
                            placeholder="Enter your password"
                            required
                            autoComplete="current-password"
                        />

                        <Button
                            type="submit"
                            className="w-full"
                            isLoading={isLoading}
                        >
                            Sign In
                        </Button>

                        <div className="relative py-4">
                            <div className="absolute inset-0 flex items-center">
                                <div className="w-full border-t border-border-subtle" />
                            </div>
                            <div className="relative flex justify-center">
                                <span className="px-4 text-caption text-text-tertiary bg-bg-elevated">
                                    New to Telegram?
                                </span>
                            </div>
                        </div>

                        <Link to="/register" className="block">
                            <Button type="button" variant="secondary" className="w-full">
                                Create an account
                            </Button>
                        </Link>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
};
