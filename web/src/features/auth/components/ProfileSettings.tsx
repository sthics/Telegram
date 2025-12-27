import { useState, useRef, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Camera } from 'lucide-react';
import { authApi } from '../api';
import { chatApi } from '@/features/chat/api';
import { useAuthStore } from '../store';
import { Button } from '@/shared/components/Button';
import { Input } from '@/shared/components/Input';
import { Modal } from '@/shared/components/Modal';
import { Skeleton, SkeletonAvatar } from '@/shared/components/Skeleton';
import { Avatar } from '@/shared/components/Avatar';

interface ProfileSettingsProps {
    isOpen: boolean;
    onClose: () => void;
}

// Profile skeleton loading component
const ProfileSkeleton = () => (
    <div className="space-y-6 animate-fade-in">
        {/* Avatar skeleton */}
        <div className="flex flex-col items-center">
            <SkeletonAvatar size="xl" className="w-24 h-24" />
            <Skeleton variant="text" width={120} className="mt-3" />
        </div>

        {/* Form fields skeleton */}
        <div className="space-y-4">
            <div className="space-y-1.5">
                <Skeleton variant="text" width={80} height={14} />
                <Skeleton variant="rectangular" height={40} className="rounded-lg" />
            </div>
            <div className="space-y-1.5">
                <Skeleton variant="text" width={40} height={14} />
                <Skeleton variant="rectangular" height={80} className="rounded-lg" />
            </div>
            <div className="space-y-1.5">
                <Skeleton variant="text" width={60} height={14} />
                <Skeleton variant="rectangular" height={40} className="rounded-lg" />
            </div>
        </div>
    </div>
);

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

    return (
        <Modal isOpen={isOpen} onClose={onClose} title="Edit Profile" size="md">
            {isLoading ? (
                <ProfileSkeleton />
            ) : (
                <div className="space-y-6 animate-fade-in">
                    {/* Avatar */}
                    <div className="flex flex-col items-center">
                        <div className="relative group">
                            <div className="w-24 h-24 rounded-full overflow-hidden">
                                {avatarPreview ? (
                                    <img
                                        src={avatarPreview}
                                        alt="Avatar"
                                        className="w-full h-full object-cover"
                                    />
                                ) : (
                                    <Avatar
                                        name={username || profile?.email || 'User'}
                                        size="xl"
                                        className="w-full h-full"
                                    />
                                )}
                            </div>
                            <button
                                onClick={() => fileInputRef.current?.click()}
                                disabled={isUploading}
                                className="absolute inset-0 rounded-full bg-bg-base/60 backdrop-blur-sm opacity-0 group-hover:opacity-100 transition-all duration-200 flex items-center justify-center cursor-pointer"
                            >
                                {isUploading ? (
                                    <div className="w-6 h-6 border-2 border-white border-t-transparent rounded-full animate-spin" />
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
                        <p className="text-caption text-text-tertiary mt-2">
                            Click to change photo
                        </p>
                    </div>

                    {/* Form fields */}
                    <div className="space-y-4">
                        <Input
                            label="Username"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                            placeholder="Enter username"
                        />

                        <div className="space-y-1.5">
                            <label className="block text-label font-medium text-text-secondary">
                                Bio
                            </label>
                            <textarea
                                value={bio}
                                onChange={(e) => setBio(e.target.value)}
                                placeholder="Write something about yourself..."
                                rows={3}
                                className="w-full px-3 py-2.5 bg-bg border border-border-default rounded-lg text-body text-text-primary placeholder:text-text-tertiary focus:outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-500/20 transition-all resize-none"
                            />
                        </div>

                        <Input
                            label="Email"
                            type="email"
                            value={profile?.email || ''}
                            disabled
                            hint="Email cannot be changed"
                        />
                    </div>

                    {/* Actions */}
                    <div className="flex justify-end gap-3 pt-2">
                        <Button variant="ghost" onClick={onClose}>
                            Cancel
                        </Button>
                        <Button
                            onClick={handleSave}
                            isLoading={updateMutation.isPending}
                        >
                            Save Changes
                        </Button>
                    </div>
                </div>
            )}
        </Modal>
    );
};
