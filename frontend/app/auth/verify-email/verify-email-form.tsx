"use client";

import { FormEvent, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { ApiError } from "@/lib/api";
import { resendVerificationEmail, verifyEmail } from "@/lib/auth";
import {
    AUTH_INPUT_CLASS_NAME,
    AuthErrorMessage,
    AuthSubmitButton,
} from "@/components/auth/auth-form-primitives";

export function VerifyEmailForm() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const initialEmail = searchParams.get("email") ?? "";
    const [email, setEmail] = useState(initialEmail);
    const [code, setCode] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [isResending, setIsResending] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");
    const [infoMessage, setInfoMessage] = useState("");

    async function onSubmit(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setErrorMessage("");
        setInfoMessage("");
        setIsLoading(true);

        try {
            await verifyEmail({ email, code });
            router.push("/auth/login?verified=1");
        } catch (error) {
            if (error instanceof ApiError) {
                setErrorMessage(error.message);
            } else {
                setErrorMessage("Unable to verify email");
            }
        } finally {
            setIsLoading(false);
        }
    }

    async function onResend() {
        setErrorMessage("");
        setInfoMessage("");
        setIsResending(true);

        try {
            await resendVerificationEmail(email);
            setInfoMessage("If the email is registered, a new code has been sent.");
        } catch (error) {
            if (error instanceof ApiError) {
                setErrorMessage(error.message);
            } else {
                setErrorMessage("Unable to resend code");
            }
        } finally {
            setIsResending(false);
        }
    }

    return (
        <form className="space-y-4" onSubmit={onSubmit}>
            <p className="text-xs text-zinc-400">
                We sent a 6-digit code to your email. Enter it below to activate
                your account.
            </p>
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
                <label
                    className="block text-xs font-medium text-zinc-200"
                    htmlFor="code"
                >
                    Verification Code
                </label>
                <input
                    id="code"
                    className={`${AUTH_INPUT_CLASS_NAME} tracking-[0.4em]`}
                    placeholder="123456"
                    required
                    inputMode="numeric"
                    autoComplete="one-time-code"
                    minLength={4}
                    maxLength={10}
                    value={code}
                    onChange={(event) =>
                        setCode(event.target.value.replace(/\s+/g, ""))
                    }
                />
            </div>
            <AuthErrorMessage message={errorMessage} />
            {infoMessage ? (
                <p className="rounded-lg bg-emerald-500/10 px-3 py-2 text-xs text-emerald-300">
                    {infoMessage}
                </p>
            ) : null}
            <AuthSubmitButton
                isLoading={isLoading}
                loadingText="Verifying..."
                idleText="Verify Email"
            />
            <button
                type="button"
                className="w-full text-center text-xs text-zinc-400 underline decoration-zinc-700 underline-offset-4 transition-colors hover:text-white hover:decoration-white disabled:opacity-50"
                onClick={onResend}
                disabled={isResending || !email}
            >
                {isResending ? "Sending..." : "Resend code"}
            </button>
        </form>
    );
}
