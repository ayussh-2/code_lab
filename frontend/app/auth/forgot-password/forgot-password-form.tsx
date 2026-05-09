"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { confirmPasswordReset, requestPasswordReset } from "@/lib/auth";
import {
    AUTH_INPUT_CLASS_NAME,
    AuthErrorMessage,
    AuthSubmitButton,
} from "@/components/auth/auth-form-primitives";

type Stage = "request" | "confirm";

export function ForgotPasswordForm() {
    const router = useRouter();
    const [stage, setStage] = useState<Stage>("request");
    const [email, setEmail] = useState("");
    const [code, setCode] = useState("");
    const [newPassword, setNewPassword] = useState("");
    const [isLoading, setIsLoading] = useState(false);
    const [errorMessage, setErrorMessage] = useState("");
    const [infoMessage, setInfoMessage] = useState("");

    async function onRequest(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setErrorMessage("");
        setInfoMessage("");
        setIsLoading(true);

        try {
            await requestPasswordReset(email);
            setInfoMessage(
                "If the email is registered, a reset code has been sent.",
            );
            setStage("confirm");
        } catch (error) {
            if (error instanceof ApiError) {
                setErrorMessage(error.message);
            } else {
                setErrorMessage("Unable to send reset code");
            }
        } finally {
            setIsLoading(false);
        }
    }

    async function onConfirm(event: FormEvent<HTMLFormElement>) {
        event.preventDefault();
        setErrorMessage("");
        setInfoMessage("");
        setIsLoading(true);

        try {
            await confirmPasswordReset({
                email,
                code,
                new_password: newPassword,
            });
            router.push("/auth/login?reset=1");
        } catch (error) {
            if (error instanceof ApiError) {
                setErrorMessage(error.message);
            } else {
                setErrorMessage("Unable to reset password");
            }
        } finally {
            setIsLoading(false);
        }
    }

    if (stage === "request") {
        return (
            <form className="space-y-4" onSubmit={onRequest}>
                <p className="text-xs text-zinc-400">
                    Enter your email and we will send you a one-time code to
                    reset your password.
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
                <AuthErrorMessage message={errorMessage} />
                <AuthSubmitButton
                    isLoading={isLoading}
                    loadingText="Sending..."
                    idleText="Send Reset Code"
                />
            </form>
        );
    }

    return (
        <form className="space-y-4" onSubmit={onConfirm}>
            {infoMessage ? (
                <p className="rounded-lg bg-emerald-500/10 px-3 py-2 text-xs text-emerald-300">
                    {infoMessage}
                </p>
            ) : null}
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
                    Reset Code
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
            <div className="space-y-1">
                <label
                    className="block text-xs font-medium text-zinc-200"
                    htmlFor="new-password"
                >
                    New Password
                </label>
                <input
                    id="new-password"
                    className={AUTH_INPUT_CLASS_NAME}
                    placeholder="At least 6 characters"
                    required
                    type="password"
                    minLength={6}
                    value={newPassword}
                    onChange={(event) => setNewPassword(event.target.value)}
                />
            </div>
            <AuthErrorMessage message={errorMessage} />
            <AuthSubmitButton
                isLoading={isLoading}
                loadingText="Resetting..."
                idleText="Reset Password"
            />
            <button
                type="button"
                className="w-full text-center text-xs text-zinc-400 underline decoration-zinc-700 underline-offset-4 transition-colors hover:text-white hover:decoration-white"
                onClick={() => setStage("request")}
            >
                Use a different email
            </button>
        </form>
    );
}
