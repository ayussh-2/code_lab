"use client";

import dynamic from "next/dynamic";
import rehypeKatex from "rehype-katex";
import remarkMath from "remark-math";
import "@uiw/react-md-editor/markdown-editor.css";
import "katex/dist/katex.min.css";

const MDEditor = dynamic(() => import("@uiw/react-md-editor"), { ssr: false });

interface ProblemMarkdownEditorProps {
  value: string;
  onChange: (value: string) => void;
  height?: number;
}

export function ProblemMarkdownEditor({
  value,
  onChange,
  height = 440,
}: ProblemMarkdownEditorProps) {
  return (
    <div
      data-color-mode="dark"
      className="admin-md-editor [&_.w-md-editor]:rounded-lg [&_.w-md-editor]:shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08)] [&_.w-md-editor]:!bg-[#141414] [&_.w-md-editor-bar]:!border-b-white/[0.08] [&_.w-md-editor-bar]:!bg-[#1a1a1a] [&_.w-md-editor-content]:!bg-[#141414] [&_.w-md-editor-preview]:!bg-[#141414] [&_.w-md-editor-text]:!bg-[#141414] [&_.w-md-editor-text-pre>code]:!bg-transparent [&_.w-md-editor-text-textarea]:!font-mono [&_.w-md-editor-text-textarea]:!text-[13px] [&_.w-md-editor-text-textarea]:!text-zinc-300"
    >
      <MDEditor
        value={value}
        height={height}
        preview="live"
        visibleDragbar
        onChange={(v) => onChange(v ?? "")}
        textareaProps={{
          placeholder:
            "Markdown statement. Inline math: $x^2$. Display: $$\\sum_{i=1}^n i$$",
        }}
        previewOptions={{
          remarkPlugins: [remarkMath],
          rehypePlugins: [rehypeKatex],
          className:
            "!bg-[#141414] !p-4 text-sm leading-7 text-zinc-400 [&_.katex]:text-zinc-200",
        }}
      />
    </div>
  );
}
