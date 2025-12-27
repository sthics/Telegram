import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { MessageCircle, AlertCircle, ArrowLeft } from 'lucide-react';
import { useAuthStore } from '@/features/auth/store';
import { authApi } from '@/features/auth/api';
import { Button } from '@/shared/components/Button';
import { Input } from '@/shared/components/Input';
import { Card, CardContent, CardHeader } from '@/shared/components/Card';

export const RegisterPage = () => {
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
        const username = formData.get('username') as string;
        const password = formData.get('password') as string;
        const firstName = formData.get('firstName') as string;
        const lastName = formData.get('lastName') as string;

        try {
            const { accessToken, user } = await authApi.register({
                email,
                username,
                password,
                firstName,
                lastName
            });
            setAuth(accessToken, user);
            navigate('/');
        } catch (err: any) {
            setError(err.response?.data?.error || 'Failed to register');
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
                <div className="absolute -top-40 -left-40 w-80 h-80 bg-brand-500/10 rounded-full blur-3xl" />
                <div className="absolute -bottom-40 -right-40 w-80 h-80 bg-brand-600/10 rounded-full blur-3xl" />
            </div>

            <Card
                className={`w-full max-w-md relative animate-fade-in ${shake ? 'animate-shake' : ''}`}
                variant="elevated"
            >
                <CardHeader className="text-center pb-2">
                    {/* Back link */}
                    <Link
                        to="/login"
                        className="absolute left-6 top-6 p-2 rounded-full hover:bg-bg-elevated transition-colors"
                    >
                        <ArrowLeft className="w-5 h-5 text-text-secondary" />
                    </Link>

                    {/* Logo */}
                    <div className="mx-auto w-16 h-16 bg-gradient-to-br from-brand-400 to-brand-600 rounded-2xl flex items-center justify-center mb-4 shadow-lg shadow-brand-500/25">
                        <MessageCircle className="w-8 h-8 text-white" />
                    </div>
                    <h1 className="text-h1 text-text-primary">Create account</h1>
                    <p className="text-body text-text-secondary mt-1">
                        Join Telegram and start messaging
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

                        <div className="grid grid-cols-2 gap-3">
                            <Input
                                name="firstName"
                                label="First name"
                                placeholder="John"
                                required
                                autoComplete="given-name"
                            />
                            <Input
                                name="lastName"
                                label="Last name"
                                placeholder="Doe"
                                autoComplete="family-name"
                            />
                        </div>

                        <Input
                            name="username"
                            label="Username"
                            placeholder="johndoe"
                            required
                            autoComplete="username"
                            hint="This will be your unique identifier"
                        />

                        <Input
                            name="email"
                            type="email"
                            label="Email"
                            placeholder="john@example.com"
                            required
                            autoComplete="email"
                        />

                        <Input
                            name="password"
                            type="password"
                            label="Password"
                            placeholder="Create a password"
                            required
                            autoComplete="new-password"
                            hint="Use 8 or more characters"
                        />

                        <Button
                            type="submit"
                            className="w-full"
                            isLoading={isLoading}
                        >
                            Create Account
                        </Button>

                        <p className="text-center text-caption text-text-tertiary">
                            By signing up, you agree to our Terms of Service and Privacy Policy
                        </p>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
};
