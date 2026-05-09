"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { login } from "@/lib/auth";
import {
    AUTH_INPUT_CLASS_NAME,
    AuthErrorMessage,
    AuthSocialDivider,
    AuthSubmitButton,
} from "@/components/auth/auth-form-primitives";
import { SocialLoginButtons } from "@/components/auth/social-login-buttons";

export function LoginForm() {
    const router = useRouter();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");

    async function onSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setErrorMessage("");
        setIsLoading(true);

        try {
            const result = await login({ email, password });
            localStorage.setItem(
                "code_lab_access_token",
                result.data.access_token,
            );
            localStorage.setItem(
                "code_lab_user",
                JSON.stringify(result.data.user),
            );
            router.push("/problems");
        } catch (error) {
            if (error instanceof ApiError) {
                if (error.status === 403 && /not verified/i.test(error.message)) {
                    router.push(
                        `/auth/verify-email?email=${encodeURIComponent(email)}`,
                    );
                    return;
                }
                setErrorMessage(error.message);
            } else {
                setErrorMessage("Unable to login");
            }
        } finally {
            setIsLoading(false);
        }
    }

    return (
        <>
            <form className="space-y-4" onSubmit={onSubmit}>
                <div className="space-y-1">
                    <label
                        className="block text-xs font-medium text-zinc-200"
                        htmlFor="email"
                    >
                        Email Address
                    </label>
                    <input
                        id="email"
                        className={AUTH_INPUT_CLASS_NAME}
                        placeholder="name@company.com"
                        required
                        type="email"
                        value={email}
                        onChange={(event) => setEmail(event.target.value)}
                    />
                </div>
                <div className="space-y-1">
                    <div className="flex items-center justify-between">
                        <label
                            className="block text-xs font-medium text-zinc-200"
                            htmlFor="password"
                        >
                            Password
                        </label>
                        <Link
                            href="/auth/forgot-password"
                            className="text-xs text-zinc-400 underline decoration-zinc-700 underline-offset-4 transition-colors hover:text-white hover:decoration-white"
                        >
                            Forgot password?
                        </Link>
                    </div>
                    <input
                        id="password"
                        className={AUTH_INPUT_CLASS_NAME}
                        placeholder="••••••••"
                        required
                        type="password"
                        minLength={6}
                        value={password}
                        onChange={(event) => setPassword(event.target.value)}
                    />
                </div>
                <AuthErrorMessage message={errorMessage} />
                <AuthSubmitButton
                    isLoading={isLoading}
                    loadingText="Signing In..."
                    idleText="Sign In"
                />
            </form>
            <AuthSocialDivider />
            <SocialLoginButtons />
        </>
    );
}
