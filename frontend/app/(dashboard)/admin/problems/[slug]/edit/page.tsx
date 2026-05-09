"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import {
    type ProblemDetail,
    getProblemBySlug,
    updateProblem,
} from "@/lib/problems";
import {
    ProblemForm,
    type ProblemFormInitialValues,
} from "@/components/admin/problem-form";
import { useAdminGuard } from "@/hooks/use-admin-guard";

function toFormValues(p: ProblemDetail): ProblemFormInitialValues {
    return {
        title: p.title,
        difficulty: p.difficulty,
        topicIds: p.topic_ids ?? [],
        hints: p.hints.length > 0 ? p.hints : [""],
        details: p.details,
        examples:
            p.examples.length > 0
                ? p.examples
                : [{ input: "", output: "", explanation: "" }],
        constraints: p.constraints.length > 0 ? p.constraints : [""],
        samples:
            p.sample_test_cases.length > 0
                ? p.sample_test_cases.map((s) => ({
                      input: s.input,
                      expected: s.expected,
                  }))
                : [{ input: "", expected: "" }],
    };
}

export default function AdminEditProblemPage() {
    const router = useRouter();
    const params = useParams<{ slug: string }>();
    const allowed = useAdminGuard();
    const [initialValues, setInitialValues] =
        useState<ProblemFormInitialValues | null>(null);
    const [loadError, setLoadError] = useState("");
    const slug = params?.slug;

    useEffect(() => {
        if (!allowed || !slug) return;
        async function load() {
            try {
                const result = await getProblemBySlug(slug as string);
                setInitialValues(toFormValues(result.data));
            } catch (error) {
                setLoadError(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load problem",
                );
            }
        }
        void load();
    }, [allowed, slug]);

    if (!allowed) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                Checking access...
            </div>
        );
    }

    if (loadError) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center bg-[#0e0e0e] text-sm text-red-400">
                {loadError}
            </div>
        );
    }

    if (!initialValues) {
        return (
            <div className="flex min-h-[50vh] items-center justify-center bg-[#0e0e0e] text-sm text-zinc-500">
                Loading problem...
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-[#0e0e0e] text-zinc-300">
            <main className="mx-auto max-w-[1100px] px-6 py-10">
                <header className="mb-10">
                    <p className="text-[11px] font-medium uppercase tracking-widest text-zinc-500">
                        Admin
                    </p>
                    <h1 className="mt-2 text-3xl font-semibold tracking-[-0.08em] text-white">
                        Edit problem
                    </h1>
                    <p className="mt-2 max-w-xl text-sm leading-7 text-zinc-500">
                        The slug stays the same so existing links and bookmarks
                        keep working. Sample test cases will be replaced with
                        whatever you save.
                    </p>
                </header>

                <ProblemForm
                    mode="edit"
                    initialValues={initialValues}
                    submitLabel="Save changes"
                    submittingLabel="Saving..."
                    onSubmit={async (payload) => {
                        await updateProblem(slug as string, payload);
                        router.push("/admin/problems");
                    }}
                />
            </main>
        </div>
    );
}
