"use client";

import { type TestCaseInput } from "@/lib/problems";

interface TestCasesEditorProps {
    title: string;
    description?: string;
    value: TestCaseInput[];
    onChange: (next: TestCaseInput[]) => void;
    addLabel?: string;
}

const labelClass =
    "mb-2 block text-[11px] font-semibold uppercase tracking-widest text-zinc-500";

const inputClass =
    "w-full rounded-md border border-white/[0.08] bg-[#1a1a1a] px-3 py-2 text-sm text-zinc-200 outline-none transition-shadow focus-visible:ring-2 focus-visible:ring-[hsla(212,100%,48%,1)]";

function emptyCase(): TestCaseInput {
    return { input: "", expected: "" };
}

export function TestCasesEditor({
    title,
    description,
    value,
    onChange,
    addLabel = "Add case",
}: TestCasesEditorProps) {
    const cases = value.length > 0 ? value : [emptyCase()];

    function update(index: number, patch: Partial<TestCaseInput>) {
        const next = cases.map((c, i) => (i === index ? { ...c, ...patch } : c));
        onChange(next);
    }

    function add() {
        onChange([...cases, emptyCase()]);
    }

    function remove(index: number) {
        const next = cases.filter((_, i) => i !== index);
        onChange(next.length > 0 ? next : [emptyCase()]);
    }

    return (
        <section className="rounded-lg border border-white/[0.08] bg-[#141414] p-6 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.04)_0px_2px_2px]">
            <div className="flex items-center justify-between gap-4">
                <h2 className="text-xs font-semibold uppercase tracking-widest text-zinc-500">
                    {title}
                </h2>
                <button
                    type="button"
                    onClick={add}
                    className="rounded-md border border-white/[0.1] bg-white/[0.04] px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/[0.08]"
                >
                    {addLabel}
                </button>
            </div>
            {description ? (
                <p className="mt-3 max-w-2xl text-xs leading-6 text-zinc-500">
                    {description}
                </p>
            ) : null}
            <div className="mt-4 space-y-6">
                {cases.map((c, i) => (
                    <div
                        key={`tc-${i}`}
                        className="space-y-3 rounded-md border border-white/[0.06] bg-[#1a1a1a] p-4"
                    >
                        <div className="flex items-center justify-between">
                            <p className="font-mono text-[11px] font-medium uppercase tracking-wide text-zinc-500">
                                Case {i + 1}
                            </p>
                            <button
                                type="button"
                                onClick={() => remove(i)}
                                className="text-[11px] font-medium text-zinc-500 transition-colors hover:text-red-400"
                            >
                                Remove
                            </button>
                        </div>
                        <div className="grid gap-3 md:grid-cols-2">
                            <div>
                                <label className={labelClass}>Input</label>
                                <textarea
                                    className={`${inputClass} min-h-[88px] resize-y font-mono text-xs`}
                                    value={c.input}
                                    onChange={(e) =>
                                        update(i, { input: e.target.value })
                                    }
                                />
                            </div>
                            <div>
                                <label className={labelClass}>Expected</label>
                                <textarea
                                    className={`${inputClass} min-h-[88px] resize-y font-mono text-xs`}
                                    value={c.expected}
                                    onChange={(e) =>
                                        update(i, { expected: e.target.value })
                                    }
                                />
                            </div>
                        </div>
                    </div>
                ))}
            </div>
        </section>
    );
}
