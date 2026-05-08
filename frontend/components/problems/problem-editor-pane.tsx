"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import dynamic from "next/dynamic";
import {
  DEFAULT_EDITOR_CODE,
  EDITOR_LANGUAGES,
  type EditorLanguage,
} from "./problem-shared";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
});

interface ProblemEditorPaneProps {
  backHref: string;
}

export function ProblemEditorPane({ backHref }: ProblemEditorPaneProps) {
  const [selectedLanguage, setSelectedLanguage] = useState<EditorLanguage>(
    EDITOR_LANGUAGES[0],
  );
  const [code, setCode] = useState(DEFAULT_EDITOR_CODE[EDITOR_LANGUAGES[0].id]);
  const [isLanguageMenuOpen, setIsLanguageMenuOpen] = useState(false);
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

  return (
    <section className="flex min-w-0 flex-1 flex-col overflow-hidden rounded-lg border border-white/8 bg-[#141414]">
      <div className="flex shrink-0 items-center justify-between border-b border-white/8 bg-[#1a1a1a] px-3 py-2">
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

        <div className="flex items-center gap-1 text-zinc-600">
          <button className="rounded p-1.5 text-xs transition-colors hover:text-zinc-300" title="Reset">
            ↺
          </button>
          <button className="rounded p-1.5 text-xs transition-colors hover:text-zinc-300" title="Settings">
            ⚙
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

      <div className="flex shrink-0 items-center justify-between border-t border-white/8 bg-[#1a1a1a] p-3">
        <Link
          href={backHref}
          className="rounded border border-white/10 bg-white/4 px-3 py-1.5 text-xs text-zinc-300 transition-colors hover:bg-white/8"
        >
          ← Back
        </Link>
        <div className="flex gap-2">
          <button className="rounded border border-white/10 bg-white/4 px-4 py-1.5 text-xs text-zinc-300 transition-colors hover:bg-white/8">
            Run Code
          </button>
          <button className="rounded bg-[#1cbf73] px-5 py-1.5 text-xs font-semibold text-black transition-opacity hover:opacity-90">
            Submit
          </button>
        </div>
      </div>
    </section>
  );
}
