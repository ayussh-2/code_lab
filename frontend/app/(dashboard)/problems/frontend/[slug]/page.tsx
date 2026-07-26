"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { ApiError } from "@/lib/api";
import { type ProblemDetail, getProblemBySlug } from "@/lib/problems";
export interface FrontendQuestion {
    title: string;
    description: string;
    requirements: string[];
}

function toFrontendQuestion(problem: ProblemDetail): FrontendQuestion {
    const requirements =
        problem.hints.length > 0 ? problem.hints : problem.constraints;
    return {
        title: problem.title,
        description: problem.details,
        requirements,
    };
}

export default function SolveFrontendProblemPage() {
    const params = useParams<{ slug: string }>();
    const [question, setQuestion] = useState<FrontendQuestion | null>(null);
    const [errorMessage, setErrorMessage] = useState("");
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        async function loadProblem() {
            setIsLoading(true);
            setErrorMessage("");
            try {
                const result = await getProblemBySlug(params.slug);
                setQuestion(toFrontendQuestion(result.data));
            } catch (error) {
                setErrorMessage(
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

    if (isLoading) {
        return (
            <div className="flex h-[calc(100dvh-49px)] items-center justify-center text-sm text-zinc-400">
                Loading problem...
            </div>
        );
    }

    if (errorMessage || !question) {
        return (
            <div className="flex h-[calc(100dvh-49px)] items-center justify-center text-sm text-red-300">
                {errorMessage || "Problem not found"}
            </div>
        );
    }

    return (
        <div className="flex h-[calc(100dvh-49px)] flex-col items-center justify-center text-zinc-400 p-8 text-center bg-zinc-950">
            <div className="max-w-md p-6 rounded-xl border border-zinc-800 bg-zinc-900/50 backdrop-blur-sm">
                <h2 className="text-xl font-semibold mb-2 text-zinc-200">{question.title}</h2>
                <p className="text-sm text-zinc-500 mb-4">Frontend coding playground is currently offline.</p>
                <div className="text-left text-xs bg-zinc-950 p-4 rounded border border-zinc-800/80 text-zinc-400 font-mono mb-4 overflow-auto max-h-40">
                    <strong className="text-zinc-300 block mb-1">Problem Description:</strong>
                    {question.description}
                </div>
                <p className="text-xs text-zinc-500">
                    We are currently redesigning the frontend practice workspace. Please try solving algorithmic problems in the standard code editor!
                </p>
            </div>
        </div>
    );
}
