"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";

export default function GoogleAuthCallbackPage() {
    const router = useRouter();
    const searchParams = useSearchParams();

    useEffect(() => {
        const accessToken = searchParams.get("access_token");
        const user = searchParams.get("user");

        if (!accessToken || !user) {
            router.replace("/login?error=google_auth_failed");
            return;
        }

        try {
            const base64 = user.replace(/-/g, "+").replace(/_/g, "/");
            const padded = base64 + "=".repeat((4 - (base64.length % 4)) % 4);
            const decoded = atob(padded);
            localStorage.setItem("code_lab_access_token", accessToken);
            localStorage.setItem("code_lab_user", decoded);
            router.replace("/problems");
        } catch {
            router.replace("/login?error=google_auth_failed");
        }
    }, [router, searchParams]);

    return (
        <div className="flex min-h-screen items-center justify-center bg-black text-sm text-zinc-300">
            Finishing Google sign in...
        </div>
    );
}
