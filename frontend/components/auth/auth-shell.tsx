import type { ReactNode } from "react";
import Link from "next/link";

interface AuthShellProps {
  title: string;
  subtitle: string;
  footerText: string;
  footerLinkText: string;
  footerLinkHref: string;
  children: ReactNode;
}

export function AuthShell({
  title,
  subtitle,
  footerText,
  footerLinkText,
  footerLinkHref,
  children,
}: AuthShellProps) {
  return (
    <main className="w-full max-w-[420px]">
      <div className="mb-10 text-center">
        <h1 className="mb-2 text-4xl font-semibold tracking-[-0.08em] text-white">
          {title}
        </h1>
        <p className="text-sm text-zinc-400">{subtitle}</p>
      </div>

      <div className="rounded-xl bg-[#0a0a0a] p-6 shadow-[rgba(255,255,255,0.08)_0px_0px_0px_1px,rgba(0,0,0,0.5)_0px_10px_20px_-12px]">
        {children}
      </div>

      <div className="mt-6 text-center text-sm text-zinc-500">
        {footerText}{" "}
        <Link
          href={footerLinkHref}
          className="text-white underline decoration-zinc-700 underline-offset-4 transition-colors hover:decoration-white"
        >
          {footerLinkText}
        </Link>
      </div>
    </main>
  );
}
