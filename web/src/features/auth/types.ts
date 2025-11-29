export interface User {
    id: number;
    username: string;
    email: string;
    firstName?: string;
    lastName?: string;
    bio?: string;
    avatarUrl?: string;
    lastSeen?: number; // Timestamp
}

export interface AuthResponse {
    accessToken: string;
    refreshToken: string;
    userId: number;
    user: User;
}

export interface LoginRequest {
    email: string;
    password: string; // Plain text, hashed on server
}

export interface RegisterRequest {
    username: string;
    email: string;
    password: string;
    firstName?: string;
    lastName?: string;
}
