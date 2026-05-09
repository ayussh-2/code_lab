"use client";

import { useEffect, useMemo, useSyncExternalStore } from "react";
import { useRouter } from "next/navigation";
import { ApiError, apiRequestWithAuth } from "@/lib/api";
import { logout } from "@/lib/auth";
import { getAccessToken, getStoredUser } from "@/lib/session";
import { TopNav } from "@/components/site/top-nav";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import { useState } from "react";

interface MeResponse {
    user_id: number;
    email: string;
    role: string;
}

interface StoredUser {
    name: string;
    email: string;
}

export default function ProfilePage() {
    const router = useRouter();
    const [me, setMe] = useState<MeResponse | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setErrorMessage] = useState("");

    const isClient = useSyncExternalStore(
        () => () => {},
        () => true,
        () => false,
    );

    const storedUser = useMemo<StoredUser | null>(() => {
        if (!isClient) return null;
        return getStoredUser<StoredUser>();
    }, [isClient]);

    useEffect(() => {
        async function loadProfile() {
            if (!isClient) return;
            try {
                const result = await apiRequestWithAuth<MeResponse>(
                    "/auth/me",
                    getAccessToken(),
                );
                setMe(result.data);
            } catch (error) {
                if (error instanceof ApiError && error.status === 401) {
                    router.replace("/auth/login");
                    return;
                }
                setErrorMessage(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load profile",
                );
            } finally {
                setIsLoading(false);
            }
        }
        void loadProfile();
    }, [isClient, router]);

    async function handleLogout() {
        try {
            await logout();
        } finally {
            router.replace("/auth/login");
        }
    }

    if (!isClient) {
        return (
            <div className="flex min-h-screen items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                <LoadingState message="Loading..." />
            </div>
        );
    }

    return (
        <div className="flex min-h-screen flex-col bg-[#0e0e0e] text-zinc-300">
            <main className="mx-auto grid w-full max-w-[1100px] flex-1 grid-cols-1 gap-4 px-6 py-10 md:grid-cols-12">
                {/* Left col */}
                <section className="col-span-12 space-y-4 md:col-span-4">
                    <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-6">
                        <div className="mb-5 flex items-center gap-4">
                            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-white/[0.06] text-2xl font-medium text-white">
                                {(storedUser?.name?.[0] ?? "U").toUpperCase()}
                            </div>
                            <div>
                                <h1 className="text-base font-semibold text-white">
                                    {storedUser?.name ?? "User"}
                                </h1>
                                <p className="text-xs text-zinc-500">
                                    {storedUser?.email ?? "-"}
                                </p>
                            </div>
                        </div>
                        {isLoading && (
                            <LoadingState
                                className="text-xs"
                                message="Loading profile..."
                            />
                        )}
                        {errorMessage && (
                            <ErrorState
                                className="text-xs"
                                message={errorMessage}
                            />
                        )}
                        <button
                            type="button"
                            onClick={handleLogout}
                            className="mt-4 w-full rounded border border-white/[0.1] bg-white/[0.04] py-2 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                        >
                            Logout
                        </button>
                    </div>
                </section>

                {/* Right col */}
                <section className="col-span-12 space-y-4 md:col-span-8">
                    <div className="grid grid-cols-2 gap-4">
                        <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-5">
                            <p className="text-[11px] uppercase tracking-widest text-zinc-500">
                                User ID
                            </p>
                            <p className="mt-2 text-3xl font-semibold text-white">
                                #{me?.user_id ?? "-"}
                            </p>
                        </div>
                        <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-5">
                            <p className="text-[11px] uppercase tracking-widest text-zinc-500">
                                Role
                            </p>
                            <p className="mt-2 text-3xl font-semibold capitalize text-white">
                                {me?.role ?? "-"}
                            </p>
                        </div>
                    </div>

                    <div className="overflow-hidden rounded-lg border border-white/[0.08] bg-[#141414]">
                        <div className="border-b border-white/[0.08] px-5 py-3">
                            <h2 className="text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                                Session
                            </h2>
                        </div>
                        <div className="divide-y divide-white/[0.06]">
                            <div className="flex items-center justify-between px-5 py-4">
                                <div>
                                    <p className="text-sm font-medium text-white">
                                        Email
                                    </p>
                                    <p className="text-xs text-zinc-500">
                                        {me?.email ?? storedUser?.email ?? "-"}
                                    </p>
                                </div>
                                <span className="rounded bg-[#1cbf73]/10 px-2 py-0.5 font-mono text-xs text-[#1cbf73]">
                                    Live
                                </span>
                            </div>
                            <div className="flex items-center justify-between px-5 py-4">
                                <div>
                                    <p className="text-sm font-medium text-white">
                                        Status
                                    </p>
                                    <p className="text-xs text-zinc-500">
                                        Authenticated via /api/auth/me
                                    </p>
                                </div>
                                <span className="rounded bg-[#1cbf73]/10 px-2 py-0.5 font-mono text-xs text-[#1cbf73]">
                                    Active
                                </span>
                            </div>
                        </div>
                    </div>
                </section>
            </main>
        </div>
    );
}
