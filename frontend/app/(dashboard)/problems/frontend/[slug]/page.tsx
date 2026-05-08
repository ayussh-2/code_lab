"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { ApiError } from "@/lib/api";
import { type ProblemDetail, getProblemBySlug } from "@/lib/problems";
import {
    FrontendWebContainer,
    type FrontendQuestion,
} from "@/components/problems/frontend-webcontainer";

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

    return <FrontendWebContainer question={question} />;
}
