"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/components/auth/auth-context";

export function useAdminGuard(): boolean | null {
    const router = useRouter();
    const { user, isLoading } = useAuth();

    useEffect(() => {
        if (isLoading) return;
        if (!user) {
            router.replace("/auth/login");
            return;
        }
        if (user.role !== "admin" && user.role !== "problem_setter") {
            router.replace("/problems");
        }
    }, [user, isLoading, router]);

    if (isLoading) return null;
    if (!user) return null;
    if (user.role !== "admin" && user.role !== "problem_setter") return null;
    return true;
}
