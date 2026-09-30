"use client";

import dynamic from "next/dynamic";
import { startTransition, useEffect, useState } from "react";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
  ssr: false,
});

const REACT_JSX_DEFINITIONS = `
declare namespace JSX {
  interface Element {}
  interface IntrinsicElements {
    [elementName: string]: any;
  }
}

declare module "react" {
  export function useState<T>(initialValue: T): [T, (value: T) => void];
  export function useEffect(effect: () => void | (() => void), deps?: unknown[]): void;
  export const StrictMode: any;
}

declare module "react-dom/client" {
  export function createRoot(container: Element | null): {
    render(element: JSX.Element): void;
  };
}
`;

interface CodeEditorPanelProps {
  filePath: string | undefined;
  readFile: (path: string) => Promise<string>;
  writeFile: (path: string, contents: string) => Promise<void>;
}

function getLanguage(filePath: string) {
  const extension = filePath.split(".").pop()?.toLowerCase();
  return (
    {
      css: "css",
      html: "html",
      js: "javascript",
      jsx: "javascript",
      json: "json",
      md: "markdown",
      ts: "typescript",
      tsx: "typescript",
    }[extension ?? ""] ?? "plaintext"
  );
}

export function CodeEditorPanel({
  filePath,
  readFile,
  writeFile,
}: CodeEditorPanelProps) {
  const [contents, setContents] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [saveState, setSaveState] = useState<
    "idle" | "saving" | "saved" | "error"
  >("idle");

  useEffect(() => {
    if (!filePath) return;
    let cancelled = false;

    startTransition(() => {
      setIsLoading(true);
      setSaveState("idle");
    });
    readFile(filePath)
      .then((nextContents) => {
        if (!cancelled) setContents(nextContents);
      })
      .catch(() => {
        if (!cancelled) {
          setContents("Unable to read this file from the WebContainer.");
          setSaveState("error");
        }
      })
      .finally(() => {
        if (!cancelled) setIsLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [filePath, readFile]);

  useEffect(() => {
    if (!filePath || isLoading) return;
    startTransition(() => setSaveState("saving"));
    const timeoutId = window.setTimeout(() => {
      writeFile(filePath, contents)
        .then(() => setSaveState("saved"))
        .catch(() => setSaveState("error"));
    }, 250);

    return () => window.clearTimeout(timeoutId);
  }, [contents, filePath, isLoading, writeFile]);

  if (!filePath) {
    return (
      <section className="flex min-w-0 flex-1 items-center justify-center bg-zinc-950 text-sm text-zinc-500">
        Select a file to start editing.
      </section>
    );
  }

  return (
    <section className="flex min-w-0 flex-1 flex-col overflow-hidden bg-[#1e1e1e]">
      <div className="flex h-10 shrink-0 items-center justify-between border-b border-zinc-800 px-3">
        <span className="truncate font-mono text-xs text-zinc-300">
          {filePath}
        </span>
        <span className="ml-3 shrink-0 text-[11px] text-zinc-500">
          {isLoading
            ? "Loading..."
            : saveState === "saving"
              ? "Saving..."
              : saveState === "saved"
                ? "Saved"
                : saveState === "error"
                  ? "Save failed"
                  : ""}
        </span>
      </div>
      <div className="min-h-0 flex-1">
        {isLoading ? (
          <div className="flex h-full items-center justify-center text-sm text-zinc-500">
            Loading file...
          </div>
        ) : (
          <MonacoEditor
            height="100%"
            language={getLanguage(filePath)}
            path={filePath}
            theme="vs-dark"
            value={contents}
            onChange={(value) => setContents(value ?? "")}
            beforeMount={(monaco) => {
              const compilerOptions = {
                allowJs: true,
                allowSyntheticDefaultImports: true,
                allowNonTsExtensions: true,
                jsx: monaco.languages.typescript.JsxEmit.ReactJSX,
                module: monaco.languages.typescript.ModuleKind.ESNext,
                moduleResolution:
                  monaco.languages.typescript.ModuleResolutionKind.NodeJs,
                noEmit: true,
                target: monaco.languages.typescript.ScriptTarget.ESNext,
              };

              monaco.languages.typescript.javascriptDefaults.setCompilerOptions(
                compilerOptions,
              );
              monaco.languages.typescript.typescriptDefaults.setCompilerOptions(
                compilerOptions,
              );
              monaco.languages.typescript.typescriptDefaults.addExtraLib(
                REACT_JSX_DEFINITIONS,
                "file:///node_modules/@types/react/index.d.ts",
              );
              monaco.languages.typescript.javascriptDefaults.addExtraLib(
                REACT_JSX_DEFINITIONS,
                "file:///node_modules/@types/react/index.d.ts",
              );
            }}
            options={{
              automaticLayout: true,
              fontSize: 13,
              minimap: { enabled: false },
              padding: { top: 12 },
              scrollBeyondLastLine: false,
            }}
          />
        )}
      </div>
    </section>
  );
}
