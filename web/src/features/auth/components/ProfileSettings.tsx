import { useState, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { X, Camera, Loader2, User as UserIcon } from 'lucide-react';
import { authApi } from '../api';
import { chatApi } from '@/features/chat/api';
import { useAuthStore } from '../store';
import { Button } from '@/shared/components/Button';

interface ProfileSettingsProps {
    isOpen: boolean;
    onClose: () => void;
}

export const ProfileSettings = ({ isOpen, onClose }: ProfileSettingsProps) => {
    const queryClient = useQueryClient();
    const setAuth = useAuthStore((state) => state.setAuth);
    const token = useAuthStore((state) => state.token);

    const [username, setUsername] = useState('');
    const [bio, setBio] = useState('');
    const [avatarPreview, setAvatarPreview] = useState<string | null>(null);
    const [isUploading, setIsUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);

    // Fetch current profile
    const { data: profile, isLoading } = useQuery({
        queryKey: ['profile'],
        queryFn: authApi.getProfile,
        enabled: isOpen,
    });

    // Sync form state when profile data loads
    useEffect(() => {
        if (profile) {
            setUsername(profile.username || '');
            setBio(profile.bio || '');
            setAvatarPreview(profile.avatar_url || null);
        }
    }, [profile]);

    // Update profile mutation
    const updateMutation = useMutation({
        mutationFn: authApi.updateProfile,
        onSuccess: (updatedUser) => {
            queryClient.setQueryData(['profile'], updatedUser);
            // Update auth store
            if (token) {
                setAuth(token, updatedUser);
            }
            onClose();
        },
    });

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const file = e.target.files?.[0];
        if (!file) return;

        // Show preview immediately
        const reader = new FileReader();
        reader.onload = (ev) => {
            setAvatarPreview(ev.target?.result as string);
        };
        reader.readAsDataURL(file);

        // Upload to MinIO
        setIsUploading(true);
        try {
            const { uploadUrl, objectKey } = await chatApi.getPresignedUrl(file.name, file.type || 'image/jpeg');
            await chatApi.uploadFileToUrl(uploadUrl, file, file.type || 'image/jpeg');
            const publicUrl = `http://localhost:9000/chat-media/${objectKey}`;
            updateMutation.mutate({ avatar_url: publicUrl });
        } catch (error) {
            console.error('Failed to upload avatar:', error);
            setAvatarPreview(profile?.avatar_url || null);
        } finally {
            setIsUploading(false);
            if (fileInputRef.current) fileInputRef.current.value = '';
        }
    };

    const handleSave = () => {
        updateMutation.mutate({
            username: username || undefined,
            bio: bio || undefined,
        });
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 animate-in fade-in duration-200">
            <div className="bg-surface rounded-2xl w-full max-w-md mx-4 shadow-2xl animate-in zoom-in-95 duration-200">
                {/* Header */}
                <div className="flex items-center justify-between p-4 border-b border-border-subtle">
                    <h2 className="text-lg font-semibold text-text-primary">Edit Profile</h2>
                    <Button size="icon" variant="ghost" onClick={onClose}>
                        <X className="w-5 h-5" />
                    </Button>
                </div>

                {/* Content */}
                <div className="p-6">
                    {isLoading ? (
                        <div className="flex justify-center py-8">
                            <Loader2 className="w-8 h-8 animate-spin text-brand-primary" />
                        </div>
                    ) : (
                        <div className="space-y-6">
                            {/* Avatar */}
                            <div className="flex flex-col items-center">
                                <div className="relative group">
                                    <div className="w-24 h-24 rounded-full bg-gradient-to-br from-brand-primary to-brand-hover flex items-center justify-center text-white text-3xl font-medium overflow-hidden">
                                        {avatarPreview ? (
                                            <img
                                                src={avatarPreview}
                                                alt="Avatar"
                                                className="w-full h-full object-cover"
                                            />
                                        ) : (
                                            <UserIcon className="w-12 h-12" />
                                        )}
                                    </div>
                                    <button
                                        onClick={() => fileInputRef.current?.click()}
                                        disabled={isUploading}
                                        className="absolute inset-0 rounded-full bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center cursor-pointer"
                                    >
                                        {isUploading ? (
                                            <Loader2 className="w-6 h-6 text-white animate-spin" />
                                        ) : (
                                            <Camera className="w-6 h-6 text-white" />
                                        )}
                                    </button>
                                    <input
                                        ref={fileInputRef}
                                        type="file"
                                        accept="image/*"
                                        className="hidden"
                                        onChange={handleFileSelect}
                                    />
                                </div>
                                <p className="text-xs text-text-tertiary mt-2">Click to change photo</p>
                            </div>

                            {/* Username */}
                            <div>
                                <label className="block text-sm font-medium text-text-secondary mb-1.5">
                                    Username
                                </label>
                                <input
                                    type="text"
                                    value={username}
                                    onChange={(e) => setUsername(e.target.value)}
                                    placeholder="Enter username"
                                    className="w-full px-4 py-2.5 bg-app border border-border-subtle rounded-lg text-text-primary placeholder:text-text-tertiary focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary transition-all"
                                />
                            </div>

                            {/* Bio */}
                            <div>
                                <label className="block text-sm font-medium text-text-secondary mb-1.5">
                                    Bio
                                </label>
                                <textarea
                                    value={bio}
                                    onChange={(e) => setBio(e.target.value)}
                                    placeholder="Write something about yourself..."
                                    rows={3}
                                    className="w-full px-4 py-2.5 bg-app border border-border-subtle rounded-lg text-text-primary placeholder:text-text-tertiary focus:outline-none focus:border-brand-primary focus:ring-1 focus:ring-brand-primary transition-all resize-none"
                                />
                            </div>

                            {/* Email (read-only) */}
                            <div>
                                <label className="block text-sm font-medium text-text-secondary mb-1.5">
                                    Email
                                </label>
                                <input
                                    type="email"
                                    value={profile?.email || ''}
                                    disabled
                                    className="w-full px-4 py-2.5 bg-app/50 border border-border-subtle rounded-lg text-text-tertiary cursor-not-allowed"
                                />
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="flex justify-end gap-3 p-4 border-t border-border-subtle">
                    <Button variant="ghost" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button
                        onClick={handleSave}
                        disabled={updateMutation.isPending || isLoading}
                    >
                        {updateMutation.isPending ? (
                            <Loader2 className="w-4 h-4 animate-spin mr-2" />
                        ) : null}
                        Save
                    </Button>
                </div>
            </div>
        </div>
    );
};
