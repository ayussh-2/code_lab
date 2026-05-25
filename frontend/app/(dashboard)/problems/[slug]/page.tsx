"use client";

import { useEffect, useState } from "react";

import { useParams } from "next/navigation";

import { ProblemDescriptionPane } from "@/components/problems/problem-description-pane";
import { ProblemEditorPane } from "@/components/problems/problem-editor-pane";
import { PROBLEM_DETAIL_TABS } from "@/components/problems/problem-shared";
import { ApiError } from "@/lib/api";
import { getProblemBySlug, type ProblemDetail } from "@/lib/problems";

type ProblemTab = (typeof PROBLEM_DETAIL_TABS)[number];

export default function SolveProblemPage() {
    const params = useParams<{ slug: string }>();
    const [problem, setProblem] = useState<ProblemDetail | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setError] = useState("");
    const [activeTab, setActiveTab] = useState<ProblemTab>("description");

    useEffect(() => {
        async function loadProblem() {
            try {
                const result = await getProblemBySlug(params.slug);
                setProblem(result.data);
            } catch (error) {
                setError(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load problem",
                );
            } finally {
                setIsLoading(false);
            }
        }
        if (params.slug) void loadProblem();
    }, [params.slug]);

    return (
        <div className="flex h-screen flex-col overflow-hidden bg-[#0e0e0e] text-zinc-300">
            <main className="flex flex-col md:flex-1 gap-3 overflow-hidden p-3 lg:flex-row">
                <ProblemDescriptionPane
                    problem={problem}
                    isLoading={isLoading}
                    errorMessage={errorMessage}
                    activeTab={activeTab}
                    onTabChange={setActiveTab}
                    slug={params.slug}
                />
                <ProblemEditorPane
                    backHref="/problems"
                    slug={params.slug}
                    samples={problem?.sample_test_cases ?? []}
                />
            </main>
        </div>
    );
}
