"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

interface SignInRequiredProps {
    title: string;
    message: string;
}

export function SignInRequired({ title, message }: SignInRequiredProps) {
    const pathname = usePathname();
    const loginHref = `/auth/login?from=${encodeURIComponent(pathname)}`;

    return (
        <div className="rounded-lg border border-white/8 bg-[#1a1a1a] p-4">
            <h3 className="text-sm font-semibold text-white">{title}</h3>
            <p className="mt-1 max-w-md text-sm leading-6 text-zinc-400">
                {message}
            </p>
            <Link
                href={loginHref}
                className="mt-4 inline-flex rounded bg-[#1cbf73] px-4 py-2 text-xs font-semibold text-black transition-opacity hover:opacity-90"
            >
                Sign in
            </Link>
        </div>
    );
}
