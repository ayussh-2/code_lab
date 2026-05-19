"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import type { User } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { updateOwnProfile } from "@/lib/profile";

interface ProfileSettingsFormProps {
    user: User;
    onSaved: (user: User) => void;
}

export function ProfileSettingsForm({ user, onSaved }: ProfileSettingsFormProps) {
    const router = useRouter();
    const [name, setName] = useState(user.name ?? "");
    const [username, setUsername] = useState(user.username ?? "");
    const [avatarUrl, setAvatarUrl] = useState(user.avatar_url ?? "");
    const [bio, setBio] = useState(user.bio ?? "");
    const [saveMessage, setSaveMessage] = useState("");
    const [saveError, setSaveError] = useState("");
    const [isSaving, setIsSaving] = useState(false);

    async function handleSave(event: FormEvent) {
        event.preventDefault();
        setIsSaving(true);
        setSaveMessage("");
        setSaveError("");
        try {
            const result = await updateOwnProfile({
                name,
                username,
                avatar_url: avatarUrl,
                bio,
            });
            const updated: User = {
                ...user,
                name: result.data.name,
                username: result.data.username,
                avatar_url: result.data.avatar_url,
                bio: result.data.bio,
                rating: result.data.rating,
            };
            onSaved(updated);
            setSaveMessage("Profile saved.");
            if (result.data.username !== user.username) {
                router.replace(`/u/${result.data.username}`);
            }
        } catch (error) {
            setSaveError(
                error instanceof ApiError
                    ? error.message
                    : "Unable to save profile",
            );
        } finally {
            setIsSaving(false);
        }
    }

    return (
        <main className="mx-auto w-full max-w-lg px-6 py-10">
            <form
                onSubmit={handleSave}
                className="space-y-4 rounded-lg border border-white/[0.08] bg-[#141414] p-6"
            >
                <h1 className="text-lg font-semibold text-white">
                    Profile settings
                </h1>
                <p className="text-xs text-zinc-500">{user.email}</p>

                <label className="block text-sm">
                    <span className="text-zinc-400">Display name</span>
                    <input
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className="mt-1 w-full rounded border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-white"
                    />
                </label>

                <label className="block text-sm">
                    <span className="text-zinc-400">Username</span>
                    <input
                        value={username}
                        onChange={(e) => setUsername(e.target.value)}
                        className="mt-1 w-full rounded border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-white"
                    />
                    <span className="mt-1 block text-xs text-zinc-600">
                        Public URL: /u/{username || "..."}
                    </span>
                </label>

                <label className="block text-sm">
                    <span className="text-zinc-400">Avatar URL</span>
                    <input
                        value={avatarUrl}
                        onChange={(e) => setAvatarUrl(e.target.value)}
                        className="mt-1 w-full rounded border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-white"
                    />
                </label>

                <label className="block text-sm">
                    <span className="text-zinc-400">Bio</span>
                    <textarea
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        rows={4}
                        maxLength={500}
                        className="mt-1 w-full rounded border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-white"
                    />
                </label>

                {saveError ? (
                    <p className="text-sm text-[#ef4743]">{saveError}</p>
                ) : null}
                {saveMessage ? (
                    <p className="text-sm text-[#1cbf73]">{saveMessage}</p>
                ) : null}

                <button
                    type="submit"
                    disabled={isSaving}
                    className="w-full rounded bg-white/[0.1] py-2 text-sm font-medium text-white hover:bg-white/[0.14] disabled:opacity-50"
                >
                    {isSaving ? "Saving..." : "Save changes"}
                </button>
            </form>
        </main>
    );
}
