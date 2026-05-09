"use client";

import {
    type ReactNode,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useState,
} from "react";
import { ApiError } from "@/lib/api";
import { logout as apiLogout, me, type User } from "@/lib/auth";

interface AuthContextValue {
    user: User | null;
    isLoading: boolean;
    refresh: () => Promise<User | null>;
    setUser: (user: User | null) => void;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    const refresh = useCallback(async (): Promise<User | null> => {
        try {
            const result = await me();
            setUser(result.data.user);
            return result.data.user;
        } catch (error) {
            if (error instanceof ApiError && error.status === 401) {
                setUser(null);
                return null;
            }
            throw error;
        }
    }, []);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            try {
                await refresh();
            } finally {
                if (!cancelled) setIsLoading(false);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [refresh]);

    const logout = useCallback(async () => {
        try {
            await apiLogout();
        } catch (error) {
            if (!(error instanceof ApiError) || error.status !== 401) throw error;
        } finally {
            setUser(null);
        }
    }, []);

    return (
        <AuthContext.Provider
            value={{ user, isLoading, refresh, setUser, logout }}
        >
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth(): AuthContextValue {
    const ctx = useContext(AuthContext);
    if (!ctx) {
        throw new Error("useAuth must be used inside AuthProvider");
    }
    return ctx;
}
