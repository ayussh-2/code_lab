"use client";

import { useEffect, useMemo, useState } from "react";
import { ApiError } from "@/lib/api";
import { type ProblemListItem, getProblems } from "@/lib/problems";
import { ProblemsListTable } from "@/components/problems/problems-list-table";

export default function ProblemsPage() {
    const [problems, setProblems] = useState<ProblemListItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);
    const [errorMessage, setErrorMessage] = useState("");
    const [searchTerm, setSearchTerm] = useState("");

    useEffect(() => {
        async function loadProblems() {
            setIsLoading(true);
            setErrorMessage("");
            try {
                const result = await getProblems();
                setProblems(result.data);
            } catch (error) {
                if (error instanceof ApiError) {
                    setErrorMessage(error.message);
                } else {
                    setErrorMessage("Unable to load problems");
                }
            } finally {
                setIsLoading(false);
            }
        }
        void loadProblems();
    }, []);

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
            <main className="mx-auto w-full flex-1 px-6 py-10">
                <ProblemsListTable
                    problems={visibleProblems}
                    isLoading={isLoading}
                    errorMessage={errorMessage}
                    searchValue={searchTerm}
                    onSearchChange={setSearchTerm}
                />
            </main>
        </div>
    );
}
