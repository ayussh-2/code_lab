"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth/auth-context";

export function ProfileMenu() {
    const router = useRouter();
    const { user, logout, isLoading } = useAuth();
    const [open, setOpen] = useState(false);
    const [isLoggingOut, setIsLoggingOut] = useState(false);
    const containerRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        if (!open) return;
        function onClick(event: MouseEvent) {
            if (!containerRef.current) return;
            if (!containerRef.current.contains(event.target as Node)) {
                setOpen(false);
            }
        }
        function onKey(event: KeyboardEvent) {
            if (event.key === "Escape") setOpen(false);
        }
        document.addEventListener("mousedown", onClick);
        document.addEventListener("keydown", onKey);
        return () => {
            document.removeEventListener("mousedown", onClick);
            document.removeEventListener("keydown", onKey);
        };
    }, [open]);

    if (isLoading) return null;

    if (!user) {
        return (
            <Link
                href="/auth/login"
                className="text-sm text-zinc-400 transition-colors hover:text-white"
            >
                Sign in
            </Link>
        );
    }

    const initial = (user.name?.[0] ?? user.email?.[0] ?? "U").toUpperCase();
    const canAccessAdmin =
        user.role === "admin" || user.role === "problem_setter";

    async function handleLogout() {
        if (isLoggingOut) return;
        setIsLoggingOut(true);
        try {
            await logout();
        } finally {
            setIsLoggingOut(false);
            setOpen(false);
            router.replace("/auth/login");
        }
    }

    return (
        <div ref={containerRef} className="relative">
            <button
                type="button"
                aria-haspopup="menu"
                aria-expanded={open}
                onClick={() => setOpen((v) => !v)}
                className="flex h-8 w-8 items-center justify-center rounded-full bg-white/[0.08] text-xs font-medium text-white transition-colors hover:bg-white/[0.14] focus:outline-none focus-visible:ring-2 focus-visible:ring-white/40"
            >
                {initial}
            </button>

            {open ? (
                <div
                    role="menu"
                    className="absolute right-0 mt-2 w-56 overflow-hidden rounded-lg border border-white/[0.08] bg-[#141414] shadow-[rgba(0,0,0,0.5)_0px_10px_24px_-12px]"
                >
                    <div className="border-b border-white/[0.08] px-3 py-3">
                        <p className="truncate text-sm font-medium text-white">
                            {user.name || "User"}
                        </p>
                        <p className="truncate text-xs text-zinc-500">
                            {user.email}
                        </p>
                    </div>
                    <div className="py-1">
                        <Link
                            href="/profile"
                            onClick={() => setOpen(false)}
                            className="block px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-white/[0.04] hover:text-white"
                            role="menuitem"
                        >
                            Profile
                        </Link>
                        {canAccessAdmin ? (
                            <Link
                                href="/admin/problems"
                                onClick={() => setOpen(false)}
                                className="block px-3 py-2 text-sm text-zinc-300 transition-colors hover:bg-white/[0.04] hover:text-white"
                                role="menuitem"
                            >
                                Admin
                            </Link>
                        ) : null}
                        <button
                            type="button"
                            onClick={handleLogout}
                            disabled={isLoggingOut}
                            className="block w-full px-3 py-2 text-left text-sm text-zinc-300 transition-colors hover:bg-white/[0.04] hover:text-white disabled:opacity-50"
                            role="menuitem"
                        >
                            {isLoggingOut ? "Signing out..." : "Logout"}
                        </button>
                    </div>
                </div>
            ) : null}
        </div>
    );
}
