"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import { signup } from "@/lib/auth";
import {
  AUTH_INPUT_CLASS_NAME,
  AuthErrorMessage,
  AuthSocialDivider,
  AuthSubmitButton,
} from "@/components/auth/auth-form-primitives";
import { SocialLoginButtons } from "@/components/auth/social-login-buttons";

export function SignupForm() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setErrorMessage("");
    setIsLoading(true);

    try {
      await signup({ name, email, password });
      router.push("/login");
    } catch (error) {
      if (error instanceof ApiError) {
        setErrorMessage(error.message);
      } else {
        setErrorMessage("Unable to create account");
      }
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <>
      <form className="space-y-4" onSubmit={onSubmit}>
        <div className="space-y-1">
          <label className="block text-xs font-medium text-zinc-200" htmlFor="name">
            Full Name
          </label>
          <input
            id="name"
            className={AUTH_INPUT_CLASS_NAME}
            placeholder="John Doe"
            required
            type="text"
            minLength={2}
            value={name}
            onChange={(event) => setName(event.target.value)}
          />
        </div>
        <div className="space-y-1">
          <label className="block text-xs font-medium text-zinc-200" htmlFor="email">
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
          <label className="block text-xs font-medium text-zinc-200" htmlFor="password">
            Password
          </label>
          <input
            id="password"
            className={AUTH_INPUT_CLASS_NAME}
            placeholder="At least 6 characters"
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
          loadingText="Creating Account..."
          idleText="Create Account"
        />
      </form>
      <AuthSocialDivider />
      <SocialLoginButtons />
    </>
  );
}
