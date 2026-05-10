"use client";

import { useEffect, useState } from "react";
import { ApiError } from "@/lib/api";
import {
    type TestCaseInput,
    getHiddenTestCases,
    replaceHiddenTestCases,
} from "@/lib/problems";
import { TestCasesEditor } from "@/components/admin/test-cases-editor";

interface HiddenTestCasesPanelProps {
    slug: string;
}

export function HiddenTestCasesPanel({ slug }: HiddenTestCasesPanelProps) {
    const [cases, setCases] = useState<TestCaseInput[]>([
        { input: "", expected: "" },
    ]);
    const [isLoading, setIsLoading] = useState(true);
    const [loadError, setLoadError] = useState("");
    const [saveError, setSaveError] = useState("");
    const [statusMessage, setStatusMessage] = useState("");
    const [isSaving, setIsSaving] = useState(false);

    useEffect(() => {
        async function load() {
            setIsLoading(true);
            setLoadError("");
            try {
                const result = await getHiddenTestCases(slug);
                const next = result.data.map((c) => ({
                    input: c.input,
                    expected: c.expected,
                }));
                setCases(next.length > 0 ? next : [{ input: "", expected: "" }]);
            } catch (error) {
                setLoadError(
                    error instanceof ApiError
                        ? error.message
                        : "Unable to load hidden test cases",
                );
            } finally {
                setIsLoading(false);
            }
        }
        void load();
    }, [slug]);

    async function handleSave() {
        setSaveError("");
        setStatusMessage("");
        const cleaned = cases
            .map((c) => ({
                input: c.input.trim(),
                expected: c.expected.trim(),
            }))
            .filter((c) => c.input && c.expected);

        setIsSaving(true);
        try {
            const result = await replaceHiddenTestCases(slug, cleaned);
            const next = result.data.map((c) => ({
                input: c.input,
                expected: c.expected,
            }));
            setCases(next.length > 0 ? next : [{ input: "", expected: "" }]);
            setStatusMessage(
                cleaned.length === 0
                    ? "Hidden test cases cleared."
                    : `Saved ${cleaned.length} hidden test case${cleaned.length === 1 ? "" : "s"}.`,
            );
        } catch (error) {
            setSaveError(
                error instanceof ApiError
                    ? error.message
                    : "Unable to save hidden test cases",
            );
        } finally {
            setIsSaving(false);
        }
    }

    if (isLoading) {
        return (
            <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 text-sm text-zinc-500">
                Loading hidden test cases...
            </div>
        );
    }

    if (loadError) {
        return (
            <div className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 text-sm text-red-400">
                {loadError}
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <TestCasesEditor
                title="Hidden test cases"
                description="Used to grade submissions. Never shown to learners. Saving replaces all existing hidden cases for this problem."
                value={cases}
                onChange={setCases}
            />

            {saveError ? (
                <p className="text-sm text-red-400">{saveError}</p>
            ) : null}
            {statusMessage ? (
                <p className="text-sm text-emerald-400">{statusMessage}</p>
            ) : null}

            <div>
                <button
                    type="button"
                    onClick={handleSave}
                    disabled={isSaving}
                    className="rounded-md bg-[#171717] px-6 py-2.5 text-sm font-medium text-white shadow-[rgb(235,235,235)_0px_0px_0px_1px] transition-opacity hover:opacity-90 disabled:opacity-50"
                >
                    {isSaving ? "Saving..." : "Save hidden test cases"}
                </button>
            </div>
        </div>
    );
}
