"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth/auth-context";
import { ProfileSettingsForm } from "@/components/profile/profile-settings-form";
import { PublicProfileView } from "@/components/profile/public-profile-view";
import { LoadingState } from "@/components/ui/async-state";

export default function ProfilePage() {
    const router = useRouter();
    const { user, isLoading, logout, refresh, setUser } = useAuth();
    const [showSettings, setShowSettings] = useState(false);

    useEffect(() => {
        if (!isLoading && !user) {
            router.replace("/auth/login");
        }
    }, [user, isLoading, router]);

    async function handleLogout() {
        try {
            await logout();
        } finally {
            router.replace("/auth/login");
        }
    }

    if (isLoading || !user) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                <LoadingState message="Loading..." />
            </div>
        );
    }

    const publicUsername = user.username || `user-${user.id}`;

    return (
        <div className="flex min-h-screen flex-col bg-[#0e0e0e] text-zinc-300">
            <div className="border-b border-white/[0.08] bg-[#141414] px-6 py-3">
                <div className="mx-auto flex max-w-[1100px] items-center justify-between gap-4">
                    <div className="flex gap-4 text-sm">
                        <button
                            type="button"
                            onClick={() => setShowSettings(false)}
                            className={
                                !showSettings
                                    ? "text-white"
                                    : "text-zinc-500 hover:text-zinc-300"
                            }
                        >
                            Overview
                        </button>
                        <button
                            type="button"
                            onClick={() => setShowSettings(true)}
                            className={
                                showSettings
                                    ? "text-white"
                                    : "text-zinc-500 hover:text-zinc-300"
                            }
                        >
                            Settings
                        </button>
                    </div>
                    <div className="flex items-center gap-3">
                        <Link
                            href={`/u/${publicUsername}`}
                            className="text-xs text-zinc-400 hover:text-white"
                        >
                            Public profile
                        </Link>
                        <button
                            type="button"
                            onClick={handleLogout}
                            className="rounded border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs text-zinc-300 hover:bg-white/[0.08]"
                        >
                            Logout
                        </button>
                    </div>
                </div>
            </div>

            {showSettings ? (
                <ProfileSettingsForm
                    key={`${user.id}-${user.username}`}
                    user={user}
                    onSaved={(updated) => {
                        setUser(updated);
                        void refresh();
                    }}
                />
            ) : (
                <PublicProfileView
                    username={publicUsername}
                    isOwner
                />
            )}
        </div>
    );
}
