export interface User {
    id: number;
    username?: string;
    email: string;
    avatar_url?: string;
    bio?: string;
    created_at?: string;
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

export interface UpdateProfileRequest {
    username?: string;
    avatar_url?: string;
    bio?: string;
}

