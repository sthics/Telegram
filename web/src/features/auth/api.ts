import { api } from '@/shared/api/client';
import type { AuthResponse, LoginRequest, RegisterRequest, User, UpdateProfileRequest } from './types';

export const authApi = {
    login: async (data: LoginRequest): Promise<AuthResponse> => {
        const response = await api.post<AuthResponse>('/auth/login', data);
        return response.data;
    },

    register: async (data: RegisterRequest): Promise<AuthResponse> => {
        const response = await api.post<AuthResponse>('/auth/register', data);
        return response.data;
    },

    getProfile: async (): Promise<User> => {
        const response = await api.get<User>('/users/me');
        return response.data;
    },

    updateProfile: async (data: UpdateProfileRequest): Promise<User> => {
        const response = await api.patch<User>('/users/me', data);
        return response.data;
    },
};
