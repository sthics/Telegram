import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/features/auth/store';
import { authApi } from '@/features/auth/api';
import { Button } from '@/shared/components/Button';
import { Input } from '@/shared/components/Input';
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/components/Card';

export const RegisterPage = () => {
    const navigate = useNavigate();
    const setAuth = useAuthStore((state) => state.setAuth);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState('');

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
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-app p-4">
            <Card className="w-full max-w-md">
                <CardHeader className="space-y-1">
                    <CardTitle className="text-2xl text-center">Create an account</CardTitle>
                    <p className="text-center text-text-secondary text-sm">
                        Enter your details to get started.
                    </p>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit} className="space-y-4">
                        {error && (
                            <div className="p-3 rounded-md bg-status-error/10 text-status-error text-sm">
                                {error}
                            </div>
                        )}
                        <div className="grid grid-cols-2 gap-4">
                            <Input name="firstName" label="First Name" placeholder="John" required />
                            <Input name="lastName" label="Last Name" placeholder="Doe" />
                        </div>
                        <Input name="username" label="Username" placeholder="johndoe" required />
                        <Input name="email" type="email" label="Email" placeholder="john@example.com" required />
                        <Input name="password" type="password" label="Password" placeholder="••••••••" required />

                        <Button type="submit" className="w-full" isLoading={isLoading}>
                            Sign Up
                        </Button>
                        <div className="text-center text-sm text-text-secondary">
                            Already have an account?{' '}
                            <Link to="/login" className="text-brand-primary hover:underline">
                                Sign in
                            </Link>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
};
