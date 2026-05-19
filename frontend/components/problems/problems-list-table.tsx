import { ChangeEvent } from "react";
import Link from "next/link";
import type { ProblemListItem } from "@/lib/problems";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import { Chip } from "@/components/ui/chip";
import {
    PROBLEM_DIFFICULTY_TEXT_CLASS,
    getProblemHref,
} from "./problem-shared";

interface ProblemsListTableProps {
    problems: ProblemListItem[];
    isLoading: boolean;
    errorMessage: string;
    searchValue: string;
    onSearchChange: (value: string) => void;
}

const STATUS_CLASS: Record<string, string> = {
    solved: "text-[#1cbf73]",
    attempted: "text-[#e8a24a]",
    unsolved: "text-zinc-600",
};

export function ProblemsListTable({
    problems,
    isLoading,
    errorMessage,
    searchValue,
    onSearchChange,
}: ProblemsListTableProps) {
    function handleSearchChange(event: ChangeEvent<HTMLInputElement>) {
        onSearchChange(event.target.value);
    }

    return (
        <section className="w-full overflow-hidden rounded-lg border border-white/8 bg-[#141414]">
            <div className="border-b border-white/8 bg-[#1a1a1a] px-6 py-3">
                <input
                    type="text"
                    value={searchValue}
                    onChange={handleSearchChange}
                    placeholder="Search problems by title"
                    className="w-full rounded-md border border-white/8 bg-black/30 px-3 py-2 text-sm text-zinc-200 placeholder:text-zinc-500 focus:border-white/20 focus:outline-none"
                />
            </div>

            <div className="grid grid-cols-[70px_1fr_72px_72px_90px] items-center gap-4 border-b border-white/8 bg-[#1a1a1a] px-6 py-3 text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                <div>SL</div>
                <div>Title</div>
                <div className="text-right">Accept</div>
                <div className="text-right">Status</div>
                <div className="text-right">Difficulty</div>
            </div>

            {isLoading ? (
                <LoadingState
                    className="px-6 py-10"
                    message="Loading problems..."
                />
            ) : errorMessage ? (
                <ErrorState className="px-6 py-10" message={errorMessage} />
            ) : (
                problems.map((problem, index) => (
                    <Link
                        key={problem.id}
                        className="grid grid-cols-[70px_1fr_72px_72px_90px] items-center gap-4 border-b border-white/6 px-6 py-3.5 text-sm last:border-b-0 hover:bg-white/3"
                        href={getProblemHref(problem)}
                    >
                        <div className="text-xs font-mono text-zinc-500">
                            {String(index + 1).padStart(2, "0")}
                        </div>
                        <div>
                            <div className="font-medium text-zinc-200 transition-colors hover:text-white">
                                {problem.title}
                            </div>
                            <div className="mt-1 flex flex-wrap gap-1">
                                {problem.topics.map((tag) => (
                                    <Chip key={tag}>{tag}</Chip>
                                ))}
                            </div>
                        </div>
                        <div className="text-right font-mono text-xs text-zinc-400">
                            {Math.round((problem.acceptance_rate ?? 0) * 1000) / 10}%
                        </div>
                        <div
                            className={`text-right text-xs capitalize ${STATUS_CLASS[problem.status ?? "unsolved"] ?? "text-zinc-600"}`}
                        >
                            {problem.status ?? "-"}
                        </div>
                        <div
                            className={`text-right font-mono text-xs ${PROBLEM_DIFFICULTY_TEXT_CLASS[problem.difficulty.toLowerCase()]}`}
                        >
                            {problem.difficulty}
                        </div>
                    </Link>
                ))
            )}
        </section>
    );
}
