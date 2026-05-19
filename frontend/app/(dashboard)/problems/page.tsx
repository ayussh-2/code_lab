"use client";

import { useEffect, useMemo, useState } from "react";
import { useAuth } from "@/components/auth/auth-context";
import { ApiError } from "@/lib/api";
import {
    type ProblemListFilters,
    type ProblemListItem,
    type ProblemStatus,
    type TopicRow,
    getProblems,
    listTopics,
} from "@/lib/problems";
import { ProblemsListTable } from "@/components/problems/problems-list-table";

export default function ProblemsPage() {
    const { user } = useAuth();
    const [problems, setProblems] = useState<ProblemListItem[]>([]);
    const [topics, setTopics] = useState<TopicRow[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setErrorMessage] = useState("");
    const [searchTerm, setSearchTerm] = useState("");
    const [difficulty, setDifficulty] = useState("");
    const [topicId, setTopicId] = useState("");
    const [status, setStatus] = useState<ProblemStatus | "">("");

    useEffect(() => {
        let cancelled = false;
        (async () => {
            try {
                const filters: ProblemListFilters = {};
                if (difficulty) filters.difficulty = difficulty;
                if (topicId) filters.topic = Number(topicId);
                if (status && user) filters.status = status;

                const result = await getProblems(searchTerm, filters);
                if (!cancelled) setProblems(result.data);
            } catch (error) {
                if (!cancelled) {
                    setErrorMessage(
                        error instanceof ApiError
                            ? error.message
                            : "Unable to load problems",
                    );
                }
            } finally {
                if (!cancelled) setIsLoading(false);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, [difficulty, topicId, status, searchTerm, user]);

    useEffect(() => {
        let cancelled = false;
        (async () => {
            try {
                const result = await listTopics();
                if (!cancelled) setTopics(result.data);
            } catch {
                if (!cancelled) setTopics([]);
            }
        })();
        return () => {
            cancelled = true;
        };
    }, []);

    function markLoading() {
        setIsLoading(true);
        setErrorMessage("");
    }

    const visibleProblems = useMemo(() => {
        const query = searchTerm.trim().toLowerCase();
        if (!query) {
            return problems;
        }

        return problems.filter((problem) => {
            const matchesTitle = problem.title.toLowerCase().includes(query);
            const matchesTopic = problem.topics.some((topic) =>
                topic.toLowerCase().includes(query),
            );
            return matchesTitle || matchesTopic;
        });
    }, [problems, searchTerm]);

    return (
        <div className="flex min-h-screen flex-col bg-[#0e0e0e] text-zinc-200">
            <main className="mx-auto w-full flex-1 space-y-4 px-6 py-10">
                <div className="flex flex-wrap items-end gap-3 rounded-lg border border-white/8 bg-[#141414] p-4">
                    <FilterSelect
                        label="Difficulty"
                        value={difficulty}
                        onChange={(v) => {
                            markLoading();
                            setDifficulty(v);
                        }}
                        options={[
                            { value: "", label: "All" },
                            { value: "easy", label: "Easy" },
                            { value: "medium", label: "Medium" },
                            { value: "hard", label: "Hard" },
                        ]}
                    />
                    <FilterSelect
                        label="Topic"
                        value={topicId}
                        onChange={(v) => {
                            markLoading();
                            setTopicId(v);
                        }}
                        options={[
                            { value: "", label: "All" },
                            ...topics.map((t) => ({
                                value: String(t.id),
                                label: t.name,
                            })),
                        ]}
                    />
                    {user ? (
                        <FilterSelect
                            label="Status"
                            value={status}
                            onChange={(v) => {
                                markLoading();
                                setStatus(v as ProblemStatus | "");
                            }}
                            options={[
                                { value: "", label: "All" },
                                { value: "solved", label: "Solved" },
                                { value: "attempted", label: "Attempted" },
                                { value: "unsolved", label: "Unsolved" },
                            ]}
                        />
                    ) : null}
                </div>

                <ProblemsListTable
                    problems={visibleProblems}
                    isLoading={isLoading}
                    errorMessage={errorMessage}
                    searchValue={searchTerm}
                    onSearchChange={(value) => {
                        markLoading();
                        setSearchTerm(value);
                    }}
                />
            </main>
        </div>
    );
}

function FilterSelect({
    label,
    value,
    onChange,
    options,
}: {
    label: string;
    value: string;
    onChange: (value: string) => void;
    options: Array<{ value: string; label: string }>;
}) {
    return (
        <label className="text-xs text-zinc-500">
            <span className="mb-1 block uppercase tracking-widest">{label}</span>
            <select
                value={value}
                onChange={(e) => onChange(e.target.value)}
                className="rounded border border-white/8 bg-[#1a1a1a] px-2 py-1.5 text-sm text-zinc-200"
            >
                {options.map((opt) => (
                    <option key={opt.value || "all"} value={opt.value}>
                        {opt.label}
                    </option>
                ))}
            </select>
        </label>
    );
}
