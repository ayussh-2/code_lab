"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import type { ProblemDetail } from "@/lib/problems";
import {
  DEFAULT_EDITOR_CODE,
  EDITOR_LANGUAGES,
  type EditorLanguage,
} from "./problem-shared";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
});

type SampleCase = ProblemDetail["sample_test_cases"][number];

interface ProblemEditorPaneProps {
  backHref: string;
  samples: SampleCase[];
}

export function ProblemEditorPane({ backHref, samples }: ProblemEditorPaneProps) {
  const [selectedLanguage, setSelectedLanguage] = useState<EditorLanguage>(
    EDITOR_LANGUAGES[0],
  );
  const [code, setCode] = useState(DEFAULT_EDITOR_CODE[EDITOR_LANGUAGES[0].id]);
  const [isLanguageMenuOpen, setIsLanguageMenuOpen] = useState(false);
  const [isTestcasesOpen, setIsTestcasesOpen] = useState(true);
  const [activeCaseIndex, setActiveCaseIndex] = useState(0);
  const languageDropdownRef = useRef<HTMLDivElement>(null);

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
    return () => document.removeEventListener("mousedown", handleOutsideClick);
  }, []);

  function selectLanguage(language: EditorLanguage) {
    setSelectedLanguage(language);
    setCode(DEFAULT_EDITOR_CODE[language.id]);
    setIsLanguageMenuOpen(false);
  }

  const safeCaseIndex = Math.min(
    activeCaseIndex,
    Math.max(0, samples.length - 1),
  );
  const activeCase = samples[safeCaseIndex];

  return (
    <section className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-white/8 bg-[#141414]">
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
            <svg className="h-3 w-3 text-zinc-500" viewBox="0 0 12 12" fill="currentColor">
              <path d="M6 8L1 3h10L6 8z" />
            </svg>
          </button>

          {isLanguageMenuOpen && (
            <div className="absolute left-0 top-full z-50 mt-1 w-44 overflow-hidden rounded-lg border border-white/8 bg-[#1e1e1e] py-1 shadow-xl">
              {EDITOR_LANGUAGES.map((language) => (
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
            className="rounded border border-white/10 bg-white/4 px-3 py-1.5 text-xs font-medium text-zinc-300 transition-colors hover:bg-white/8"
          >
            Run
          </button>
          <button
            type="button"
            className="rounded bg-[#1cbf73] px-4 py-1.5 text-xs font-semibold text-black transition-opacity hover:opacity-90"
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
            fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
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

      <div className="shrink-0 border-t border-white/8 bg-[#141414]">
        <button
          type="button"
          onClick={() => setIsTestcasesOpen((open) => !open)}
          aria-expanded={isTestcasesOpen}
          className="flex w-full items-center justify-between px-3 py-2 text-xs font-semibold text-zinc-300 transition-colors hover:text-white"
        >
          <span className="flex items-center gap-2">
            <svg
              className={`h-3 w-3 text-zinc-500 transition-transform ${
                isTestcasesOpen ? "" : "-rotate-90"
              }`}
              viewBox="0 0 12 12"
              fill="currentColor"
            >
              <path d="M6 8L1 3h10L6 8z" />
            </svg>
            Testcase
          </span>
          <span className="font-mono text-[11px] text-zinc-500">
            {samples.length} sample{samples.length === 1 ? "" : "s"}
          </span>
        </button>

        {isTestcasesOpen && (
          <div className="max-h-56 overflow-y-auto border-t border-white/8 px-3 py-3">
            {samples.length === 0 ? (
              <p className="text-xs text-zinc-500">
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
                    <TestcaseField label="Input" value={activeCase.input} />
                    <TestcaseField
                      label="Expected"
                      value={activeCase.expected}
                    />
                  </div>
                ) : null}
              </>
            )}
          </div>
        )}
      </div>
    </section>
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
