"use client";

import type { Components } from "react-markdown";
import ReactMarkdown from "react-markdown";
import rehypeKatex from "rehype-katex";
import remarkMath from "remark-math";
import "katex/dist/katex.min.css";

const markdownComponents: Components = {
  h1: ({ children }) => (
    <h1 className="mb-4 text-2xl font-semibold tracking-[-0.06em] text-white">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="mb-3 mt-8 text-xl font-semibold tracking-[-0.05em] text-white first:mt-0">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="mb-2 mt-6 text-lg font-semibold tracking-[-0.04em] text-white first:mt-0">
      {children}
    </h3>
  ),
  p: ({ children }) => (
    <p className="mb-3 text-sm font-normal leading-7 text-zinc-400 last:mb-0">
      {children}
    </p>
  ),
  a: ({ href, children }) => (
    <a
      href={href}
      className="font-medium text-[#0072f5] underline decoration-[#0072f5]/40 underline-offset-2 hover:decoration-[#0072f5]"
      rel="noreferrer"
      target={href?.startsWith("http") ? "_blank" : undefined}
    >
      {children}
    </a>
  ),
  ul: ({ children }) => (
    <ul className="mb-4 list-disc space-y-1 pl-5 text-sm leading-7 text-zinc-400">
      {children}
    </ul>
  ),
  ol: ({ children }) => (
    <ol className="mb-4 list-decimal space-y-1 pl-5 text-sm leading-7 text-zinc-400">
      {children}
    </ol>
  ),
  li: ({ children }) => <li className="pl-0.5">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="mb-4 border-l-2 border-white/15 pl-4 text-sm text-zinc-500">
      {children}
    </blockquote>
  ),
  hr: () => <hr className="my-8 border-0 shadow-[0px_1px_0_0_rgba(255,255,255,0.08)]" />,
  pre: ({ children }) => (
    <pre className="my-4 overflow-x-auto rounded-lg bg-[#1a1a1a] p-4 font-mono text-[0.8125rem] leading-relaxed text-zinc-300 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.08),rgba(0,0,0,0.2)_0px_2px_2px]">
      {children}
    </pre>
  ),
  code: ({ className, children, ...props }) => {
    const isBlock = Boolean(className?.includes("language-"));
    if (isBlock) {
      return (
        <code className={className} {...props}>
          {children}
        </code>
      );
    }
    return (
      <code
        className="rounded-sm bg-white/[0.08] px-1.5 py-0.5 font-mono text-[0.8125rem] text-zinc-200 shadow-[0px_0px_0px_1px_rgba(255,255,255,0.06)]"
        {...props}
      >
        {children}
      </code>
    );
  },
  strong: ({ children }) => (
    <strong className="font-semibold text-zinc-200">{children}</strong>
  ),
};

interface MarkdownKatexViewProps {
  markdown: string;
  className?: string;
}

export function MarkdownKatexView({ markdown, className = "" }: MarkdownKatexViewProps) {
  return (
    <div
      className={`markdown-katex-body [&_.katex]:text-zinc-200 [&_.katex-display]:my-4 ${className}`}
    >
      <ReactMarkdown
        remarkPlugins={[remarkMath]}
        rehypePlugins={[rehypeKatex]}
        components={markdownComponents}
      >
        {markdown}
      </ReactMarkdown>
    </div>
  );
}
