"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import type { User } from "@/lib/auth";
import { getAccessToken, getStoredUser } from "@/lib/session";

export function useAdminGuard(): boolean | null {
    const router = useRouter();
    const [allowed, setAllowed] = useState<boolean | null>(null);

    useEffect(() => {
        const token = getAccessToken();
        const user = getStoredUser<User>();
        if (!token) {
            router.replace("/auth/login");
            return;
        }
        if (user?.role !== "admin" && user?.role !== "problem_setter") {
            router.replace("/problems");
            return;
        }
        setAllowed(true);
    }, [router]);

    return allowed;
}
