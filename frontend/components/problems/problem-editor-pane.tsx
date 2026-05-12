"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import { ApiError } from "@/lib/api";
import { useAuth } from "@/components/auth/auth-context";
import { SignInRequired } from "@/components/auth/sign-in-required";
import { LoadingState } from "@/components/ui/async-state";
import type { ProblemDetail } from "@/lib/problems";
import {
    createSubmission,
    type SubmissionKind,
} from "@/lib/submissions";
import { useSubmissionPoll } from "@/hooks/use-submission-poll";
import { SubmissionResult } from "./submission-result";
import {
    DEFAULT_EDITOR_CODE,
    EDITOR_LANGUAGES,
    isSubmissionSupported,
    type EditorLanguage,
} from "./problem-shared";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
    ssr: false,
});

type SampleCase = ProblemDetail["sample_test_cases"][number];

type BottomTab = "testcase" | "result";

interface ProblemEditorPaneProps {
    backHref: string;
    slug: string;
    samples: SampleCase[];
}

export function ProblemEditorPane({
    backHref,
    slug,
    samples,
}: ProblemEditorPaneProps) {
    const [selectedLanguage, setSelectedLanguage] = useState<EditorLanguage>(
        EDITOR_LANGUAGES[0],
    );
    const [code, setCode] = useState(
        DEFAULT_EDITOR_CODE[EDITOR_LANGUAGES[0].id],
    );
    const [isLanguageMenuOpen, setIsLanguageMenuOpen] = useState(false);
    const [isBottomOpen, setIsBottomOpen] = useState(true);
    const [bottomTab, setBottomTab] = useState<BottomTab>("testcase");
    const [activeCaseIndex, setActiveCaseIndex] = useState(0);
    const [submissionId, setSubmissionId] = useState<number | null>(null);
    const [submitError, setSubmitError] = useState("");
    const [isCreating, setIsCreating] = useState(false);
    const [bottomHeight, setBottomHeight] = useState(320);
    const [isResizing, setIsResizing] = useState(false);
    const languageDropdownRef = useRef<HTMLDivElement>(null);
    const sectionRef = useRef<HTMLElement>(null);
    const { user, isLoading: isAuthLoading } = useAuth();

    const {
        submission,
        isPolling,
        errorMessage: pollError,
    } = useSubmissionPoll(submissionId);

    useEffect(() => {
        function handleOutsideClick(event: MouseEvent) {
            if (
                languageDropdownRef.current &&
                !languageDropdownRef.current.contains(event.target as Node)
            ) {
                setIsLanguageMenuOpen(false);
            }
        }
        document.addEventListener("mousedown", handleOutsideClick);
        return () =>
            document.removeEventListener("mousedown", handleOutsideClick);
    }, []);

    useEffect(() => {
        if (!isResizing) return;

        function handleMove(event: MouseEvent) {
            if (!sectionRef.current) return;
            const rect = sectionRef.current.getBoundingClientRect();
            const next = rect.bottom - event.clientY;
            const min = 140;
            const max = Math.max(min, rect.height - 160);
            setBottomHeight(Math.min(max, Math.max(min, next)));
        }

        function handleUp() {
            setIsResizing(false);
        }

        document.addEventListener("mousemove", handleMove);
        document.addEventListener("mouseup", handleUp);
        const previousCursor = document.body.style.cursor;
        const previousSelect = document.body.style.userSelect;
        document.body.style.cursor = "row-resize";
        document.body.style.userSelect = "none";

        return () => {
            document.removeEventListener("mousemove", handleMove);
            document.removeEventListener("mouseup", handleUp);
            document.body.style.cursor = previousCursor;
            document.body.style.userSelect = previousSelect;
        };
    }, [isResizing]);

    function selectLanguage(language: EditorLanguage) {
        setSelectedLanguage(language);
        setCode(DEFAULT_EDITOR_CODE[language.id]);
        setIsLanguageMenuOpen(false);
    }

    async function dispatchSubmission(kind: SubmissionKind) {
        setSubmitError("");

        if (isAuthLoading) return;
        if (!user) {
            setSubmissionId(null);
            setBottomTab("result");
            setIsBottomOpen(true);
            return;
        }

        if (!isSubmissionSupported(selectedLanguage.id)) {
            setSubmitError(
                `${selectedLanguage.label} isn't supported by the judge yet. Try Python, JavaScript, or C++.`,
            );
            setBottomTab("result");
            setIsBottomOpen(true);
            return;
        }

        setIsCreating(true);
        setBottomTab("result");
        setIsBottomOpen(true);

        try {
            const result = await createSubmission({
                problem_slug: slug,
                language: selectedLanguage.id,
                source_code: code,
                kind,
            });
            setSubmissionId(result.data.submission_id);
        } catch (error) {
            setSubmitError(
                error instanceof ApiError
                    ? error.message
                    : "Unable to submit code",
            );
        } finally {
            setIsCreating(false);
        }
    }

    const isBusy = isCreating || isPolling || isAuthLoading;
    const safeCaseIndex = Math.min(
        activeCaseIndex,
        Math.max(0, samples.length - 1),
    );
    const activeCase = samples[safeCaseIndex];

    return (
        <section
            ref={sectionRef}
            className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-white/8 bg-[#141414]"
        >
            <div className="flex shrink-0 items-center gap-2 border-b border-white/8 bg-[#1a1a1a] px-3 py-2">
                <Link
                    href={backHref}
                    aria-label="Back to problems"
                    className="flex h-7 w-7 items-center justify-center rounded border border-white/10 bg-white/4 text-xs text-zinc-400 transition-colors hover:bg-white/8 hover:text-zinc-200"
                >
                    ←
                </Link>

                <div className="relative" ref={languageDropdownRef}>
                    <button
                        onClick={() => setIsLanguageMenuOpen((open) => !open)}
                        className="flex items-center gap-1.5 rounded border border-white/10 bg-white/4 px-3 py-1.5 font-mono text-xs text-zinc-300 transition-colors hover:bg-white/8"
                    >
                        {selectedLanguage.label}
                        <svg
                            className="h-3 w-3 text-zinc-500"
                            viewBox="0 0 12 12"
                            fill="currentColor"
                        >
                            <path d="M6 8L1 3h10L6 8z" />
                        </svg>
                    </button>

                    {isLanguageMenuOpen && (
                        <div className="absolute left-0 top-full z-50 mt-1 w-44 overflow-hidden rounded-lg border border-white/8 bg-[#1e1e1e] py-1 shadow-xl">
                            {EDITOR_LANGUAGES.filter((language) =>
                                isSubmissionSupported(language.id),
                            ).map((language) => (
                                <button
                                    key={language.id}
                                    onClick={() => selectLanguage(language)}
                                    className={`w-full px-3 py-2 text-left font-mono text-xs transition-colors hover:bg-white/6 ${
                                        language.id === selectedLanguage.id
                                            ? "text-white"
                                            : "text-zinc-400"
                                    }`}
                                >
                                    {language.label}
                                </button>
                            ))}
                        </div>
                    )}
                </div>

                <div className="ml-auto flex items-center gap-2">
                    <button
                        type="button"
                        onClick={() => dispatchSubmission("run")}
                        disabled={isBusy}
                        className="rounded border border-white/10 bg-white/4 px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/8 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        Run
                    </button>
                    <button
                        type="button"
                        onClick={() => dispatchSubmission("submit")}
                        disabled={isBusy}
                        className="rounded bg-[#1cbf73] px-4 py-1.5 text-xs font-semibold text-black transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-50"
                    >
                        Submit
                    </button>
                </div>
            </div>

            <div className="flex-1 overflow-hidden">
                <MonacoEditor
                    height="100%"
                    language={selectedLanguage.id}
                    value={code}
                    onChange={(value) => setCode(value ?? "")}
                    theme="vs-dark"
                    options={{
                        fontSize: 13,
                        fontFamily:
                            "'JetBrains Mono', 'Fira Code', monospace",
                        fontLigatures: true,
                        minimap: { enabled: false },
                        scrollBeyondLastLine: false,
                        lineNumbers: "on",
                        renderLineHighlight: "line",
                        tabSize: 4,
                        padding: { top: 12, bottom: 12 },
                        overviewRulerLanes: 0,
                        hideCursorInOverviewRuler: true,
                        scrollbar: {
                            vertical: "auto",
                            horizontal: "auto",
                            verticalScrollbarSize: 6,
                            horizontalScrollbarSize: 6,
                        },
                    }}
                />
            </div>

            <div
                className="flex shrink-0 flex-col border-t border-white/8 bg-[#141414]"
                style={{ height: isBottomOpen ? bottomHeight : undefined }}
            >
                {isBottomOpen && (
                    <div
                        role="separator"
                        aria-orientation="horizontal"
                        onMouseDown={(event) => {
                            event.preventDefault();
                            setIsResizing(true);
                        }}
                        onDoubleClick={() => setBottomHeight(320)}
                        className={`group flex h-1.5 shrink-0 cursor-row-resize items-center justify-center transition-colors ${
                            isResizing
                                ? "bg-[#1cbf73]/40"
                                : "bg-white/5 hover:bg-white/15"
                        }`}
                    >
                        <span className="h-0.5 w-10 rounded-full bg-white/20 group-hover:bg-white/40" />
                    </div>
                )}

                <div className="sticky top-0 z-10 flex shrink-0 items-center justify-between bg-[#141414] px-3 py-2">
                    <div className="flex items-center gap-1">
                        <button
                            type="button"
                            onClick={() => setIsBottomOpen((open) => !open)}
                            aria-expanded={isBottomOpen}
                            aria-label={
                                isBottomOpen ? "Collapse panel" : "Expand panel"
                            }
                            className="flex items-center gap-2 rounded p-1 text-xs font-semibold text-zinc-300 transition-colors hover:bg-white/6 hover:text-white"
                        >
                            <svg
                                className={`h-3 w-3 text-zinc-500 transition-transform ${
                                    isBottomOpen ? "" : "-rotate-90"
                                }`}
                                viewBox="0 0 12 12"
                                fill="currentColor"
                            >
                                <path d="M6 8L1 3h10L6 8z" />
                            </svg>
                        </button>
                        <BottomTabButton
                            label="Testcase"
                            active={bottomTab === "testcase"}
                            onClick={() => {
                                setBottomTab("testcase");
                                setIsBottomOpen(true);
                            }}
                        />
                        <BottomTabButton
                            label="Result"
                            active={bottomTab === "result"}
                            onClick={() => {
                                setBottomTab("result");
                                setIsBottomOpen(true);
                            }}
                        />
                    </div>

                    {bottomTab === "testcase" && isBottomOpen ? (
                        <span className="font-mono text-[11px] text-zinc-500">
                            {samples.length} sample
                            {samples.length === 1 ? "" : "s"}
                        </span>
                    ) : null}
                </div>

                {isBottomOpen && (
                    <div className="min-h-0 flex-1 overflow-y-auto border-t border-white/8 px-3 py-3">
                        {bottomTab === "testcase" ? (
                            <TestcaseTabBody
                                samples={samples}
                                activeCase={activeCase}
                                safeCaseIndex={safeCaseIndex}
                                setActiveCaseIndex={setActiveCaseIndex}
                            />
                        ) : isAuthLoading ? (
                            <LoadingState message="Checking sign in..." />
                        ) : !user ? (
                            <SignInRequired
                                title="Sign in to run code"
                                message="You need to be signed in before you can run code, submit solutions, or view judge results."
                            />
                        ) : (
                            <SubmissionResult
                                submission={submission}
                                isPolling={isCreating || isPolling}
                                errorMessage={submitError || pollError}
                            />
                        )}
                    </div>
                )}
            </div>
        </section>
    );
}

function BottomTabButton({
    label,
    active,
    onClick,
}: {
    label: string;
    active: boolean;
    onClick: () => void;
}) {
    return (
        <button
            type="button"
            onClick={onClick}
            className={`rounded px-2 py-1 text-xs font-medium transition-colors ${
                active
                    ? "bg-white/8 text-white"
                    : "text-zinc-500 hover:text-zinc-300"
            }`}
        >
            {label}
        </button>
    );
}

function TestcaseTabBody({
    samples,
    activeCase,
    safeCaseIndex,
    setActiveCaseIndex,
}: {
    samples: SampleCase[];
    activeCase: SampleCase | undefined;
    safeCaseIndex: number;
    setActiveCaseIndex: (index: number) => void;
}) {
    return (
        <>
            <IOHint />
            {samples.length === 0 ? (
                <p className="mt-2 text-xs text-zinc-500">
                    No sample test cases for this problem.
                </p>
            ) : (
                <>
                    <div className="mb-3 flex flex-wrap gap-2">
                        {samples.map((_, index) => (
                            <button
                                key={index}
                                type="button"
                                onClick={() => setActiveCaseIndex(index)}
                                className={`rounded px-2.5 py-1 font-mono text-[11px] transition-colors ${
                                    index === safeCaseIndex
                                        ? "bg-white/10 text-white"
                                        : "bg-white/4 text-zinc-400 hover:bg-white/8 hover:text-zinc-200"
                                }`}
                            >
                                Case {index + 1}
                            </button>
                        ))}
                    </div>
                    {activeCase ? (
                        <div className="space-y-3">
                            <TestcaseField
                                label="Input (stdin)"
                                value={activeCase.input}
                            />
                            <TestcaseField
                                label="Expected (stdout)"
                                value={activeCase.expected}
                            />
                        </div>
                    ) : null}
                </>
            )}
        </>
    );
}

function IOHint() {
    return (
        <div className="mb-3 rounded border border-white/8 bg-[#1a1a1a] px-3 py-2 text-[11px] leading-5 text-zinc-400">
            <span className="font-semibold text-zinc-300">I/O format:</span>{" "}
            your program reads from <span className="font-mono text-zinc-300">stdin</span>{" "}
            and prints the answer to{" "}
            <span className="font-mono text-zinc-300">stdout</span>. Match the
            exact format shown in the sample input below.
        </div>
    );
}

function TestcaseField({ label, value }: { label: string; value: string }) {
    return (
        <div>
            <p className="mb-1 text-[10px] font-semibold uppercase tracking-widest text-zinc-500">
                {label}
            </p>
            <pre className="whitespace-pre-wrap rounded border border-white/8 bg-[#1a1a1a] px-3 py-2 font-mono text-xs leading-6 text-zinc-300">
                {value}
            </pre>
        </div>
    );
}
