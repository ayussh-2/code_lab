"use client";

import type { ReactNode } from "react";
import { useEffect, useState } from "react";

import Link from "next/link";

import { getProblemHref } from "@/components/problems/problem-shared";
import { ActivityHeatmap } from "@/components/profile/activity-heatmap";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import { ApiError } from "@/lib/api";
import {
    type ActivityDay,
    getActivityHeatmap,
    getProfileSubmissions,
    getPublicProfile,
    getUserStats,
    type ProfileSubmission,
    type PublicProfile,
    type UserStats,
} from "@/lib/profile";

const VERDICT_CLASS: Record<string, string> = {
    AC: "text-[#1cbf73]",
    WA: "text-[#ef4743]",
    TLE: "text-[#e8a24a]",
    CE: "text-zinc-400",
    RE: "text-[#ef4743]",
};

interface PublicProfileViewProps {
    username: string;
    isOwner?: boolean;
}

export function PublicProfileView({
    username,
    isOwner = false,
}: PublicProfileViewProps) {
    const [profile, setProfile] = useState<PublicProfile | null>(null);
    const [stats, setStats] = useState<UserStats | null>(null);
    const [activity, setActivity] = useState<ActivityDay[]>([]);
    const [submissions, setSubmissions] = useState<ProfileSubmission[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setErrorMessage] = useState("");

    useEffect(() => {
        async function load() {
            setIsLoading(true);
            setErrorMessage("");
            try {
                const [p, s, a, subs] = await Promise.all([
                    getPublicProfile(username),
                    getUserStats(username),
                    getActivityHeatmap(username),
                    getProfileSubmissions(username),
                ]);
                setProfile(p.data);
                setStats(s.data);
                setActivity(a.data);
                setSubmissions(subs.data);
            } catch (error) {
                setErrorMessage(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load profile",
                );
            } finally {
                setIsLoading(false);
            }
        }
        void load();
    }, [username]);

    if (isLoading) {
        return <LoadingState message="Loading profile..." />;
    }

    if (errorMessage || !profile) {
        return <ErrorState message={errorMessage || "Profile not found"} />;
    }

    const initial = (
        profile.name?.[0] ??
        profile.username[0] ??
        "U"
    ).toUpperCase();
    const acceptancePct = stats
        ? Math.round(stats.acceptance_rate * 1000) / 10
        : 0;

    return (
        <div className="mx-auto grid w-full max-w-[1100px] grid-cols-1 gap-4 px-6 py-10 md:grid-cols-12">
            <section className="col-span-12 space-y-4 md:col-span-4">
                <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-6">
                    <div className="mb-5 flex items-center gap-4">
                        {profile.avatar_url ? (
                            <img
                                src={profile.avatar_url}
                                alt=""
                                className="h-16 w-16 rounded-full object-cover"
                            />
                        ) : (
                            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-white/[0.06] text-2xl font-medium text-white">
                                {initial}
                            </div>
                        )}
                        <div>
                            <h1 className="text-base font-semibold text-white">
                                {profile.name || profile.username}
                            </h1>
                            <p className="text-xs text-zinc-500">
                                @{profile.username}
                            </p>
                        </div>
                    </div>
                    {profile.bio ? (
                        <p className="text-sm leading-relaxed text-zinc-400">
                            {profile.bio}
                        </p>
                    ) : (
                        <p className="text-sm text-zinc-600">No bio yet.</p>
                    )}
                    {isOwner ? (
                        <Link
                            href="/profile"
                            className="mt-4 inline-block text-xs text-zinc-400 underline hover:text-white"
                        >
                            Edit profile settings
                        </Link>
                    ) : null}
                </div>

                {stats ? (
                    <div className="grid grid-cols-2 gap-3">
                        <StatCard
                            label="Solved"
                            value={String(stats.total_solved)}
                        />
                        <StatCard
                            label="Easy"
                            value={String(stats.solved_by_difficulty.easy)}
                        />
                        <StatCard
                            label="Medium"
                            value={String(stats.solved_by_difficulty.medium)}
                        />
                        <StatCard
                            label="Hard"
                            value={String(stats.solved_by_difficulty.hard)}
                        />
                        <StatCard
                            label="Accept %"
                            value={`${acceptancePct}%`}
                        />
                    </div>
                ) : null}
            </section>

            <section className="col-span-12 space-y-4 md:col-span-8">
                <Panel title="Submission activity">
                    <ActivityHeatmap days={activity} />
                </Panel>

                <Panel title="Submission history">
                    {submissions.length === 0 ? (
                        <p className="text-sm text-zinc-500">
                            No submissions yet.
                        </p>
                    ) : (
                        <div className="divide-y divide-white/[0.06]">
                            {submissions.map((sub) => (
                                <Link
                                    key={sub.id}
                                    href={getProblemHref({
                                        slug: sub.problem_slug,
                                        topics: [],
                                    })}
                                    className="flex items-center justify-between gap-4 py-3 text-sm transition-colors hover:bg-white/[0.02]"
                                >
                                    <div>
                                        <p className="font-medium text-zinc-200">
                                            {sub.problem_title}
                                        </p>
                                        <p className="text-xs text-zinc-500">
                                            {sub.language} ·{" "}
                                            {new Date(
                                                sub.created_at,
                                            ).toLocaleString()}
                                        </p>
                                    </div>
                                    <span
                                        className={`font-mono text-xs font-semibold ${VERDICT_CLASS[sub.verdict] ?? "text-zinc-400"}`}
                                    >
                                        {sub.verdict}
                                    </span>
                                </Link>
                            ))}
                        </div>
                    )}
                </Panel>
            </section>
        </div>
    );
}

function StatCard({ label, value }: { label: string; value: string }) {
    return (
        <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-4">
            <p className="text-[10px] uppercase tracking-widest text-zinc-500">
                {label}
            </p>
            <p className="mt-1 text-xl font-semibold text-white">{value}</p>
        </div>
    );
}

function Panel({ title, children }: { title: string; children: ReactNode }) {
    return (
        <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-5">
            <h2 className="mb-4 text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                {title}
            </h2>
            {children}
        </div>
    );
}
