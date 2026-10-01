import { MarkdownKatexView } from "@/components/markdown/markdown-katex-view";
import type { FrontendQuestion } from "./types";

export function QuestionPanel({ question }: { question: FrontendQuestion }) {
  return (
    <div className="min-h-0 flex-1 overflow-y-auto">
      <article className="px-5 py-6">
        <MarkdownKatexView
          markdown={`# ${question.title}\n\n${question.description}`}
        />
        <h2 className="mt-8 border-t border-zinc-800 pt-6 text-lg font-semibold text-white">
          Requirements
        </h2>
        <MarkdownKatexView
          className="mt-4"
          markdown={question.requirements.map((item) => `- ${item}`).join("\n")}
        />
      </article>
    </div>
  );
}
