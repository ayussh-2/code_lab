import type { ProblemDetail } from "@/lib/problems";
import { MarkdownKatexView } from "@/components/markdown/markdown-katex-view";
import { ErrorState, LoadingState } from "@/components/ui/async-state";
import {
  PROBLEM_DETAIL_TABS,
  PROBLEM_DIFFICULTY_BADGE_CLASS,
} from "./problem-shared";

type ProblemTab = (typeof PROBLEM_DETAIL_TABS)[number];

interface ProblemDescriptionPaneProps {
  problem: ProblemDetail | null;
  isLoading: boolean;
  errorMessage: string;
  activeTab: ProblemTab;
  onTabChange: (tab: ProblemTab) => void;
}

export function ProblemDescriptionPane({
  problem,
  isLoading,
  errorMessage,
  activeTab,
  onTabChange,
}: ProblemDescriptionPaneProps) {
  return (
    <section className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-white/8 bg-[#141414]">
      <div className="flex shrink-0 items-center gap-0 border-b border-white/8 bg-[#1a1a1a] px-2">
        {PROBLEM_DETAIL_TABS.map((tab) => (
          <button
            key={tab}
            onClick={() => onTabChange(tab)}
            className={`px-3 py-3 text-xs font-medium capitalize transition-colors ${
              activeTab === tab
                ? "border-b-2 border-white text-white"
                : "text-zinc-500 hover:text-zinc-300"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {isLoading ? (
          <LoadingState message="Loading problem..." />
        ) : errorMessage ? (
          <ErrorState message={errorMessage} />
        ) : problem ? (
          <div className="space-y-5">
            <div>
              <h1 className="text-xl font-semibold tracking-tight text-white">{problem.title}</h1>
              <span
                className={`mt-2 inline-block rounded px-2 py-0.5 text-xs font-semibold uppercase ${
                  PROBLEM_DIFFICULTY_BADGE_CLASS[problem.difficulty.toLowerCase()]
                }`}
              >
                {problem.difficulty}
              </span>
            </div>

            <MarkdownKatexView markdown={problem.details} />

            {problem.examples.length > 0 && (
              <div>
                <h2 className="mb-3 text-sm font-semibold text-white">Examples</h2>
                <div className="space-y-3">
                  {problem.examples.map((example, index) => (
                    <div
                      key={`${example.input}-${index}`}
                      className="rounded-md border border-white/8 bg-[#1a1a1a] p-4 font-mono text-xs leading-6"
                    >
                      <p>
                        <span className="text-zinc-400">Input:</span>{" "}
                        <span className="text-zinc-200">{example.input}</span>
                      </p>
                      <p>
                        <span className="text-zinc-400">Output:</span>{" "}
                        <span className="text-zinc-200">{example.output}</span>
                      </p>
                      {example.explanation && (
                        <p className="mt-1 text-zinc-500">
                          <span className="text-zinc-400">Explanation:</span> {example.explanation}
                        </p>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            )}

            {problem.constraints.length > 0 && (
              <div>
                <h2 className="mb-2 text-sm font-semibold text-white">Constraints</h2>
                <MarkdownKatexView
                  markdown={problem.constraints.map((c) => `- ${c}`).join("\n")}
                />
              </div>
            )}
          </div>
        ) : null}
      </div>
    </section>
  );
}
