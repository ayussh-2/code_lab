"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import {
    type ProblemListItem,
    deleteProblem,
    getProblems,
} from "@/lib/problems";
import { getAccessToken } from "@/lib/session";
import { useAdminGuard } from "@/hooks/use-admin-guard";
import { Chip } from "@/components/ui/chip";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import { PROBLEM_DIFFICULTY_TEXT_CLASS } from "@/components/problems/problem-shared";

export default function AdminProblemsPage() {
    const allowed = useAdminGuard();
    const [problems, setProblems] = useState<ProblemListItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setErrorMessage] = useState("");
    const [pendingSlug, setPendingSlug] = useState<string | null>(null);

    useEffect(() => {
        if (!allowed) return;
        async function load() {
            setIsLoading(true);
            setErrorMessage("");
            try {
                const result = await getProblems();
                setProblems(result.data);
            } catch (error) {
                setErrorMessage(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load problems",
                );
            } finally {
                setIsLoading(false);
            }
        }
        void load();
    }, [allowed]);

    async function handleDelete(slug: string, title: string) {
        const confirmed = window.confirm(
            `Delete "${title}"? This will permanently remove the problem and its sample test cases.`,
        );
        if (!confirmed) return;

        setPendingSlug(slug);
        try {
            await deleteProblem(getAccessToken(), slug);
            setProblems((prev) => prev.filter((p) => p.slug !== slug));
        } catch (error) {
            window.alert(
                error instanceof ApiError
                    ? error.message
                    : "Unable to delete problem",
            );
        } finally {
            setPendingSlug(null);
        }
    }

    if (!allowed) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                Checking access...
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-[#0e0e0e] text-zinc-300">
            <main className="mx-auto max-w-[1100px] px-6 py-10">
                <header className="mb-8 flex flex-wrap items-end justify-between gap-4">
                    <div>
                        <p className="text-[11px] font-medium uppercase tracking-widest text-zinc-500">
                            Admin
                        </p>
                        <h1 className="mt-2 text-3xl font-semibold tracking-[-0.08em] text-white">
                            Problems
                        </h1>
                    </div>
                    <Link
                        href="/admin/problems/new"
                        className="rounded-md bg-[#171717] px-5 py-2.5 text-sm font-medium text-white shadow-[rgb(235,235,235)_0px_0px_0px_1px] transition-opacity hover:opacity-90"
                    >
                        New problem
                    </Link>
                </header>

                <section className="overflow-hidden rounded-lg border border-white/[0.08] bg-[#141414]">
                    <div className="grid grid-cols-[60px_1fr_120px_180px] items-center gap-4 border-b border-white/[0.08] bg-[#1a1a1a] px-6 py-3 text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                        <div>SL</div>
                        <div>Title</div>
                        <div>Difficulty</div>
                        <div className="text-right">Actions</div>
                    </div>

                    {isLoading ? (
                        <LoadingState
                            className="px-6 py-10"
                            message="Loading problems..."
                        />
                    ) : errorMessage ? (
                        <ErrorState
                            className="px-6 py-10"
                            message={errorMessage}
                        />
                    ) : problems.length === 0 ? (
                        <div className="px-6 py-10 text-center text-sm text-zinc-500">
                            No problems yet.
                        </div>
                    ) : (
                        problems.map((problem, index) => (
                            <div
                                key={problem.slug}
                                className="grid grid-cols-[60px_1fr_120px_180px] items-center gap-4 border-b border-white/[0.06] px-6 py-3.5 text-sm last:border-b-0"
                            >
                                <div className="font-mono text-xs text-zinc-500">
                                    {String(index + 1).padStart(2, "0")}
                                </div>
                                <div>
                                    <div className="font-medium text-zinc-200">
                                        {problem.title}
                                    </div>
                                    <div className="mt-1 flex flex-wrap gap-1">
                                        {problem.topics.map((tag) => (
                                            <Chip key={tag}>{tag}</Chip>
                                        ))}
                                    </div>
                                </div>
                                <div
                                    className={`font-mono text-xs ${PROBLEM_DIFFICULTY_TEXT_CLASS[problem.difficulty.toLowerCase()]}`}
                                >
                                    {problem.difficulty}
                                </div>
                                <div className="flex justify-end gap-2">
                                    <Link
                                        href={`/admin/problems/${problem.slug}/edit`}
                                        className="rounded-md border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                                    >
                                        Edit
                                    </Link>
                                    <button
                                        type="button"
                                        disabled={pendingSlug === problem.slug}
                                        onClick={() =>
                                            handleDelete(
                                                problem.slug,
                                                problem.title,
                                            )
                                        }
                                        className="rounded-md border border-red-500/30 bg-red-500/[0.06] px-3 py-1.5 text-xs font-medium text-red-300 transition-colors hover:bg-red-500/[0.12] disabled:opacity-50"
                                    >
                                        {pendingSlug === problem.slug
                                            ? "Deleting..."
                                            : "Delete"}
                                    </button>
                                </div>
                            </div>
                        ))
                    )}
                </section>
            </main>
        </div>
    );
}
