"use client";

import { type FormEvent, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ApiError } from "@/lib/api";
import {
    type CreateProblemPayload,
    type TopicRow,
    listTopics,
} from "@/lib/problems";
import { ProblemMarkdownEditor } from "@/components/admin/problem-markdown-editor";
import { TestCasesEditor } from "@/components/admin/test-cases-editor";

const DEFAULT_DETAILS = `## Statement

Return the sum of two integers $a$ and $b$.

$$
\\sum_{i=1}^{n} i = \\frac{n(n+1)}{2}
$$
`;

export interface ProblemFormInitialValues {
    title: string;
    difficulty: "easy" | "medium" | "hard";
    topicIds: number[];
    hints: string[];
    details: string;
    editorial: string;
    examples: Array<{ input: string; output: string; explanation: string }>;
    constraints: string[];
    samples: Array<{ input: string; expected: string }>;
}

interface ProblemFormProps {
    mode: "create" | "edit";
    initialValues?: ProblemFormInitialValues;
    submitLabel: string;
    submittingLabel: string;
    onSubmit: (payload: CreateProblemPayload) => Promise<void>;
}

const inputClass =
    "w-full rounded-md border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-sm text-zinc-200 outline-none transition-shadow focus-visible:ring-2 focus-visible:ring-[hsla(212,100%,48%,1)]";

const labelClass =
    "mb-2 block text-[11px] font-semibold uppercase tracking-widest text-zinc-500";

function emptyExample() {
    return { input: "", output: "", explanation: "" };
}

function defaults(): ProblemFormInitialValues {
    return {
        title: "",
        difficulty: "medium",
        topicIds: [],
        hints: [""],
        details: DEFAULT_DETAILS,
        editorial: "",
        examples: [emptyExample()],
        constraints: [""],
        samples: [{ input: "", expected: "" }],
    };
}

export function ProblemForm({
    initialValues,
    submitLabel,
    submittingLabel,
    onSubmit,
}: ProblemFormProps) {
    const router = useRouter();
    const seed = initialValues ?? defaults();

    const [topics, setTopics] = useState<TopicRow[]>([]);
    const [topicsError, setTopicsError] = useState("");

    const [title, setTitle] = useState(seed.title);
    const [difficulty, setDifficulty] = useState<"easy" | "medium" | "hard">(
        seed.difficulty,
    );
    const [selectedTopicIds, setSelectedTopicIds] = useState<number[]>(
        seed.topicIds,
    );
    const [hints, setHints] = useState<string[]>(
        seed.hints.length > 0 ? seed.hints : [""],
    );
    const [details, setDetails] = useState(seed.details);
    const [editorial, setEditorial] = useState(seed.editorial);
    const [examples, setExamples] = useState(
        seed.examples.length > 0 ? seed.examples : [emptyExample()],
    );
    const [constraints, setConstraints] = useState<string[]>(
        seed.constraints.length > 0 ? seed.constraints : [""],
    );
    const [samples, setSamples] = useState(
        seed.samples.length > 0 ? seed.samples : [{ input: "", expected: "" }],
    );

    const [submitError, setSubmitError] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    useEffect(() => {
        async function loadTopics() {
            try {
                const result = await listTopics();
                setTopics(result.data);
            } catch (error) {
                setTopicsError(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load topics",
                );
            }
        }
        void loadTopics();
    }, []);

    function toggleTopic(id: number) {
        setSelectedTopicIds((prev) =>
            prev.includes(id) ? prev.filter((t) => t !== id) : [...prev, id],
        );
    }

    async function handleSubmit(event: FormEvent) {
        event.preventDefault();
        setSubmitError("");

        const trimmedHints = hints.map((h) => h.trim()).filter(Boolean);
        const trimmedConstraints = constraints
            .map((c) => c.trim())
            .filter(Boolean);
        const filledExamples = examples
            .map((e) => ({
                input: e.input.trim(),
                output: e.output.trim(),
                explanation: e.explanation.trim(),
            }))
            .filter((e) => e.input && e.output);
        const filledSamples = samples
            .map((s) => ({
                input: s.input.trim(),
                expected: s.expected.trim(),
            }))
            .filter((s) => s.input && s.expected);

        if (!title.trim()) return setSubmitError("Title is required.");
        if (selectedTopicIds.length < 1)
            return setSubmitError("Select at least one topic.");
        if (trimmedHints.length < 1)
            return setSubmitError("Add at least one hint.");
        if (!details.trim())
            return setSubmitError("Problem statement is required.");
        if (filledExamples.length < 1)
            return setSubmitError("Add at least one example with input and output.");
        if (trimmedConstraints.length < 1)
            return setSubmitError("Add at least one constraint.");
        if (filledSamples.length < 1)
            return setSubmitError("Add at least one sample test case.");

        setIsSubmitting(true);
        try {
            await onSubmit({
                title: title.trim(),
                difficulty,
                topics: selectedTopicIds,
                hints: trimmedHints,
                details: details.trim(),
                editorial: editorial.trim(),
                examples: filledExamples,
                constraints: trimmedConstraints,
                sample_test_cases: filledSamples,
            });
        } catch (error) {
            setSubmitError(
                error instanceof ApiError ? error.message : "Unable to save problem",
            );
        } finally {
            setIsSubmitting(false);
        }
    }

    return (
        <form className="space-y-8" onSubmit={handleSubmit}>
            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                    Metadata
                </h2>
                <div className="mt-6 grid gap-6 md:grid-cols-2">
                    <div className="md:col-span-2">
                        <label className={labelClass} htmlFor="title">
                            Title
                        </label>
                        <input
                            id="title"
                            className={inputClass}
                            value={title}
                            onChange={(e) => setTitle(e.target.value)}
                            required
                        />
                    </div>
                    <div>
                        <label className={labelClass} htmlFor="difficulty">
                            Difficulty
                        </label>
                        <select
                            id="difficulty"
                            className={`${inputClass} cursor-pointer`}
                            value={difficulty}
                            onChange={(e) =>
                                setDifficulty(
                                    e.target.value as
                                        | "easy"
                                        | "medium"
                                        | "hard",
                                )
                            }
                        >
                            <option value="easy">Easy</option>
                            <option value="medium">Medium</option>
                            <option value="hard">Hard</option>
                        </select>
                    </div>
                </div>

                <div className="mt-6">
                    <span className={labelClass}>Topics</span>
                    {topicsError ? (
                        <p className="text-sm text-red-400">{topicsError}</p>
                    ) : (
                        <div className="mt-2 flex max-h-40 flex-wrap gap-2 overflow-y-auto rounded-md border border-white/[0.08] bg-[#1a1a1a] p-3">
                            {topics.length === 0 ? (
                                <p className="text-sm text-zinc-500">
                                    No topics yet.
                                </p>
                            ) : (
                                topics.map((t) => (
                                    <label
                                        key={t.id}
                                        className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-sm text-zinc-300 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08)] hover:bg-white/[0.04]"
                                    >
                                        <input
                                            type="checkbox"
                                            checked={selectedTopicIds.includes(
                                                t.id,
                                            )}
                                            onChange={() => toggleTopic(t.id)}
                                            className="rounded border-white/20 bg-[#141414] text-[#0072f5] focus:ring-[hsla(212,100%,48%,1)]"
                                        />
                                        {t.name}
                                    </label>
                                ))
                            )}
                        </div>
                    )}
                </div>
            </section>

            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                    Statement (Markdown + KaTeX)
                </h2>
                <div className="mt-6">
                    <ProblemMarkdownEditor
                        value={details}
                        onChange={setDetails}
                    />
                </div>
            </section>

            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                    Editorial (locked until solved)
                </h2>
                <div className="mt-6">
                    <ProblemMarkdownEditor
                        value={editorial}
                        onChange={setEditorial}
                    />
                </div>
            </section>

            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                        Hints
                    </h2>
                    <button
                        type="button"
                        onClick={() => setHints((h) => [...h, ""])}
                        className="rounded-md border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                    >
                        Add hint
                    </button>
                </div>
                <div className="mt-4 space-y-3">
                    {hints.map((hint, i) => (
                        <input
                            key={`hint-${i}`}
                            className={inputClass}
                            value={hint}
                            placeholder={`Hint ${i + 1}`}
                            onChange={(e) => {
                                const next = [...hints];
                                next[i] = e.target.value;
                                setHints(next);
                            }}
                        />
                    ))}
                </div>
            </section>

            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                        Examples
                    </h2>
                    <button
                        type="button"
                        onClick={() =>
                            setExamples((ex) => [...ex, emptyExample()])
                        }
                        className="rounded-md border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                    >
                        Add example
                    </button>
                </div>
                <div className="mt-4 space-y-6">
                    {examples.map((ex, i) => (
                        <div
                            key={`ex-${i}`}
                            className="space-y-3 rounded-md border border-white/[0.06] bg-[#1a1a1a] p-4"
                        >
                            <p className="font-mono text-[11px] font-medium uppercase tracking-wide text-zinc-500">
                                Example {i + 1}
                            </p>
                            <input
                                className={inputClass}
                                placeholder="Input"
                                value={ex.input}
                                onChange={(e) => {
                                    const next = [...examples];
                                    next[i] = {
                                        ...next[i],
                                        input: e.target.value,
                                    };
                                    setExamples(next);
                                }}
                            />
                            <input
                                className={inputClass}
                                placeholder="Output"
                                value={ex.output}
                                onChange={(e) => {
                                    const next = [...examples];
                                    next[i] = {
                                        ...next[i],
                                        output: e.target.value,
                                    };
                                    setExamples(next);
                                }}
                            />
                            <input
                                className={inputClass}
                                placeholder="Explanation (optional)"
                                value={ex.explanation}
                                onChange={(e) => {
                                    const next = [...examples];
                                    next[i] = {
                                        ...next[i],
                                        explanation: e.target.value,
                                    };
                                    setExamples(next);
                                }}
                            />
                        </div>
                    ))}
                </div>
            </section>

            <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
                <div className="flex items-center justify-between gap-4">
                    <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                        Constraints
                    </h2>
                    <button
                        type="button"
                        onClick={() => setConstraints((c) => [...c, ""])}
                        className="rounded-md border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                    >
                        Add constraint
                    </button>
                </div>
                <div className="mt-4 space-y-3">
                    {constraints.map((c, i) => (
                        <input
                            key={`c-${i}`}
                            className={inputClass}
                            value={c}
                            placeholder={`Constraint ${i + 1}`}
                            onChange={(e) => {
                                const next = [...constraints];
                                next[i] = e.target.value;
                                setConstraints(next);
                            }}
                        />
                    ))}
                </div>
            </section>

            <TestCasesEditor
                title="Sample test cases"
                description="Shown to learners on the problem page so they can verify their solution before submitting."
                value={samples}
                onChange={setSamples}
            />

            {submitError ? (
                <p className="text-sm text-red-400">{submitError}</p>
            ) : null}

            <div className="flex flex-wrap items-center gap-3">
                <button
                    type="submit"
                    disabled={isSubmitting}
                    className="rounded-md bg-[#171717] px-6 py-2.5 text-sm font-medium text-white shadow-[rgb(235,235,235)_0px_0px_0px_1px] transition-opacity hover:opacity-90 disabled:opacity-50"
                >
                    {isSubmitting ? submittingLabel : submitLabel}
                </button>
                <button
                    type="button"
                    onClick={() => router.push("/admin/problems")}
                    className="rounded-md border border-white/[0.1] bg-white/[0.04] px-5 py-2.5 text-sm font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                >
                    Cancel
                </button>
            </div>
        </form>
    );
}
