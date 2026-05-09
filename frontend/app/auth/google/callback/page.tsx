"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

export default function GoogleAuthCallbackPage() {
    const router = useRouter();

    useEffect(() => {
        router.replace("/problems");
    }, [router]);

    return (
        <div className="flex min-h-screen items-center justify-center bg-black text-sm text-zinc-300">
            Finishing Google sign in...
        </div>
    );
}
