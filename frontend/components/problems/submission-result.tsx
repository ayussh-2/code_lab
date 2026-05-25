"use client";

import {
    type Submission,
    type SubmissionTestResult,
    VERDICT_BADGE_CLASS,
    VERDICT_LABEL,
} from "@/lib/submissions";

interface SubmissionResultProps {
    submission: Submission | null;
    isPolling: boolean;
    errorMessage: string;
}

export function SubmissionResult({
    submission,
    isPolling,
    errorMessage,
}: SubmissionResultProps) {
    if (errorMessage) {
        return <p className="text-xs text-red-400">{errorMessage}</p>;
    }

    if (!submission) {
        return (
            <p className="text-xs text-zinc-500">
                {isPolling
                    ? "Submitting..."
                    : "Run or Submit to see the result."}
            </p>
        );
    }

    const verdict = submission.verdict;
    const status = submission.status;
    const showSpinner = status !== "done";

    return (
        <div className="space-y-3">
            <div className="flex flex-wrap items-center gap-3">
                <span
                    className={`rounded px-2 py-0.5 text-xs font-semibold uppercase ${VERDICT_BADGE_CLASS[verdict]}`}
                >
                    {showSpinner ? status : VERDICT_LABEL[verdict]}
                </span>

                {!showSpinner && submission.runtime_ms > 0 ? (
                    <Metric
                        label="Runtime"
                        value={`${submission.runtime_ms} ms`}
                    />
                ) : null}
                {!showSpinner && submission.memory_kb > 0 ? (
                    <Metric
                        label="Memory"
                        value={formatMemory(submission.memory_kb)}
                    />
                ) : null}
                <Metric label="Lang" value={submission.language} />
                <Metric label="Kind" value={submission.kind} />
            </div>

            {showSpinner ? (
                <p className="text-xs text-zinc-500">
                    Judging your submission, this usually takes a few seconds...
                </p>
            ) : null}

            {!showSpinner &&
            submission.results &&
            submission.results.length > 0 ? (
                <TestResultsBreakdown results={submission.results} />
            ) : null}

            {verdict === "CE" && submission.compiler_output ? (
                <CollapsibleResultBlock
                    label="Compiler output"
                    value={submission.compiler_output}
                />
            ) : null}

            {verdict === "WA" ? (
                <div className="space-y-2">
                    {submission.failed_test_case_id ? (
                        <p className="text-xs text-zinc-400">
                            Failed on test case #
                            {submission.failed_test_case_id}
                        </p>
                    ) : null}
                    <div className="grid gap-2 md:grid-cols-3">
                        <ResultBlock
                            label="Input"
                            value={submission.failed_input_preview ?? ""}
                        />
                        <ResultBlock
                            label="Expected"
                            value={submission.failed_expected_preview ?? ""}
                        />
                        <ResultBlock
                            label="Your output"
                            value={submission.failed_actual_preview ?? ""}
                        />
                    </div>
                </div>
            ) : null}

            {(verdict === "TLE" || verdict === "MLE") &&
            submission.failed_test_case_id ? (
                <p className="text-xs text-zinc-400">
                    Failed on test case #{submission.failed_test_case_id}
                </p>
            ) : null}

            {verdict === "RE" && submission.stderr_preview ? (
                <CollapsibleResultBlock
                    label="Stderr"
                    value={submission.stderr_preview}
                />
            ) : null}

            {verdict === "IE" && submission.error ? (
                <ResultBlock label="Error" value={submission.error} />
            ) : null}
        </div>
    );
}

function Metric({ label, value }: { label: string; value: string }) {
    return (
        <span className="font-mono text-[11px] text-zinc-500">
            <span className="text-zinc-600">{label}:</span>{" "}
            <span className="text-zinc-300">{value}</span>
        </span>
    );
}

function formatMemory(memoryKB: number): string {
    if (memoryKB <= 0) return "—";
    if (memoryKB > 1000) {
        return `${(memoryKB / 1024).toFixed(2)} MB`;
    }
    return `${memoryKB} KB`;
}

function ResultBlock({ label, value }: { label: string; value: string }) {
    return (
        <div>
            <p className="mb-1 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                {label}
            </p>
            <pre className="max-h-40 overflow-auto whitespace-pre-wrap rounded border border-white/8 bg-[#1a1a1a] px-3 py-2 font-mono text-xs leading-6 text-zinc-300">
                {value}
            </pre>
        </div>
    );
}

function CollapsibleResultBlock({
    label,
    value,
}: {
    label: string;
    value: string;
}) {
    return (
        <details className="rounded border border-white/8 bg-[#1a1a1a]">
            <summary className="cursor-pointer px-3 py-2 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                {label}
            </summary>
            <pre className="max-h-52 overflow-auto whitespace-pre-wrap border-t border-white/8 px-3 py-2 font-mono text-xs leading-6 text-zinc-300">
                {value}
            </pre>
        </details>
    );
}

function TestResultsBreakdown({
    results,
}: {
    results: SubmissionTestResult[];
}) {
    return (
        <details className="rounded border border-white/8 bg-[#1a1a1a]">
            <summary className="cursor-pointer px-3 py-2 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                Test Results ({results.length})
            </summary>
            <div className="border-t border-white/8 px-3 py-2">
                <div className="space-y-1">
                    {results.map((result) => (
                        <div
                            key={result.id}
                            className="grid grid-cols-[0.3fr_1fr_1fr_1fr] gap-2 text-[10px] text-zinc-400"
                        >
                            <span className="font-semibold text-zinc-500">
                                #{result.test_case_id}
                            </span>
                            <span
                                className={`inline-flex w-fit rounded px-1.5 py-0.5 text-[9px] font-semibold uppercase ${
                                    result.verdict === "AC"
                                        ? "bg-green-500/15 text-green-400"
                                        : "bg-red-500/15 text-red-400"
                                }`}
                            >
                                {result.verdict}
                            </span>
                            <span className="font-mono">
                                {result.runtime_ms}ms
                            </span>
                            <span className="font-mono">
                                {formatMemory(result.memory_kb)}
                            </span>
                        </div>
                    ))}
                </div>
            </div>
        </details>
    );
}
