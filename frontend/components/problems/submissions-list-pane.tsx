"use client";

import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/components/auth/auth-context";
import { SignInRequired } from "@/components/auth/sign-in-required";
import {
    listSubmissions,
    type Submission,
    VERDICT_BADGE_CLASS,
    VERDICT_LABEL,
} from "@/lib/submissions";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import { EDITOR_LANGUAGES } from "./problem-shared";
import { SubmissionResult } from "./submission-result";

const POLL_INTERVAL_MS = 2500;

interface SubmissionsListPaneProps {
    slug: string;
}

export function SubmissionsListPane({ slug }: SubmissionsListPaneProps) {
    const [submissions, setSubmissions] = useState<Submission[] | null>(null);
    const [errorMessage, setErrorMessage] = useState("");
    const [isUnauthorized, setIsUnauthorized] = useState(false);
    const [expandedId, setExpandedId] = useState<number | null>(null);
    const { user, isLoading: isAuthLoading } = useAuth();
    const userId = user?.id;

    useEffect(() => {
        if (isAuthLoading) return;
        if (!userId) return;

        let cancelled = false;
        let timer: ReturnType<typeof setTimeout> | null = null;

        async function load() {
            try {
                const result = await listSubmissions({
                    problem_slug: slug,
                    kind: "submit",
                });
                if (cancelled) return;
                setSubmissions(result.data);
                setErrorMessage("");
                setIsUnauthorized(false);
                const hasPending = result.data.some(
                    (s) => s.status !== "done",
                );
                if (hasPending) {
                    timer = setTimeout(load, POLL_INTERVAL_MS);
                }
            } catch (error) {
                if (cancelled) return;
                if (error instanceof ApiError && error.status === 401) {
                    setSubmissions(null);
                    setErrorMessage("");
                    setIsUnauthorized(true);
                    return;
                }
                setErrorMessage(
                    error instanceof ApiError
                        ? error.message
                        : "Cannot load submissions",
                );
            }
        }

        void load();
        return () => {
            cancelled = true;
            if (timer !== null) clearTimeout(timer);
        };
    }, [slug, userId, isAuthLoading]);

    if (isAuthLoading) return <LoadingState message="Checking sign in..." />;
    if (isUnauthorized || !user) {
        return (
            <SignInRequired
                title="Sign in to view submissions"
                message="Your submissions are private to your account. Sign in to see previous attempts and their results."
            />
        );
    }
    if (errorMessage) return <ErrorState message={errorMessage} />;
    if (submissions === null)
        return <LoadingState message="Loading submissions..." />;
    if (submissions.length === 0) {
        return (
            <p className="text-sm text-zinc-500">
                You haven&apos;t submitted anything for this problem yet.
            </p>
        );
    }

    return (
        <div className="space-y-2">
            <div className="grid grid-cols-[1.4fr_0.9fr_0.7fr_0.7fr_0.9fr] gap-3 border-b border-white/8 px-3 pb-2 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                <span>Status</span>
                <span>Language</span>
                <span>Runtime</span>
                <span>Memory</span>
                <span className="text-right">Submitted</span>
            </div>
            {submissions.map((s) => (
                <SubmissionRow
                    key={s.id}
                    submission={s}
                    expanded={expandedId === s.id}
                    onToggle={() =>
                        setExpandedId((current) =>
                            current === s.id ? null : s.id,
                        )
                    }
                />
            ))}
        </div>
    );
}

interface SubmissionRowProps {
    submission: Submission;
    expanded: boolean;
    onToggle: () => void;
}

function SubmissionRow({ submission, expanded, onToggle }: SubmissionRowProps) {
    const isTerminal = submission.status === "done";
    const verdict = submission.verdict;
    const statusLabel = isTerminal ? VERDICT_LABEL[verdict] : submission.status;

    return (
        <div className="overflow-hidden rounded-md border border-white/8 bg-[#171717]">
            <button
                type="button"
                onClick={onToggle}
                className="grid w-full grid-cols-[1.4fr_0.9fr_0.7fr_0.7fr_0.9fr] items-center gap-3 px-3 py-2 text-left text-xs transition-colors hover:bg-white/4"
            >
                <span
                    className={`inline-flex w-fit rounded px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${VERDICT_BADGE_CLASS[verdict]}`}
                >
                    {statusLabel}
                </span>
                <span className="font-mono text-[11px] text-zinc-300">
                    {formatLanguage(submission.language)}
                </span>
                <span className="font-mono text-[11px] text-zinc-400">
                    {isTerminal && submission.runtime_ms >= 0
                        ? `${submission.runtime_ms} ms`
                        : "—"}
                </span>
                <span className="font-mono text-[11px] text-zinc-400">
                    {isTerminal && submission.memory_kb > 0
                        ? `${submission.memory_kb} KB`
                        : "—"}
                </span>
                <span className="text-right font-mono text-[11px] text-zinc-500">
                    {formatRelative(submission.created_at)}
                </span>
            </button>

            {expanded ? (
                <div className="space-y-3 border-t border-white/8 bg-[#141414] px-3 py-3">
                    <SubmissionResult
                        submission={submission}
                        isPolling={!isTerminal}
                        errorMessage=""
                    />
                    <CodeBlock
                        language={submission.language}
                        code={submission.source_code}
                    />
                </div>
            ) : null}
        </div>
    );
}

function CodeBlock({ language, code }: { language: string; code: string }) {
    return (
        <div>
            <p className="mb-1 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                Code ({formatLanguage(language)})
            </p>
            <pre className="max-h-72 overflow-auto whitespace-pre rounded border border-white/8 bg-[#1a1a1a] px-3 py-2 font-mono text-xs leading-6 text-zinc-300">
                {code}
            </pre>
        </div>
    );
}

function formatLanguage(id: string): string {
    const match = EDITOR_LANGUAGES.find((l) => l.id === id);
    return match ? match.label : id;
}

function formatRelative(iso: string): string {
    const then = new Date(iso).getTime();
    if (Number.isNaN(then)) return "—";
    const diffSec = Math.max(0, Math.floor((Date.now() - then) / 1000));
    if (diffSec < 5) return "just now";
    if (diffSec < 60) return `${diffSec}s ago`;
    const min = Math.floor(diffSec / 60);
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const day = Math.floor(hr / 24);
    if (day < 7) return `${day}d ago`;
    return new Date(iso).toLocaleDateString();
}
