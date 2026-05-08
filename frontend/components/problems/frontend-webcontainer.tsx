"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import dynamic from "next/dynamic";
import type { FileSystemTree, WebContainer } from "@webcontainer/api";
import { FRONTEND_QUESTION_PROJECT } from "@/lib/frontend-question-project";

const MonacoEditor = dynamic(() => import("@monaco-editor/react"), {
    ssr: false,
});

const ANSI_ESCAPE_RE = /\x1b\[[0-9;?]*[ -/]*[@-~]/g;

function sanitizeOutput(raw: string): string {
    const stripped = raw.replace(ANSI_ESCAPE_RE, "");
    return stripped
        .split("\n")
        .map((line) => {
            const idx = line.lastIndexOf("\r");
            return idx === -1 ? line : line.slice(idx + 1);
        })
        .join("\n");
}

function flattenTreeToFiles(
    tree: FileSystemTree,
    prefix = "",
): Record<string, string> {
    const result: Record<string, string> = {};
    for (const [name, node] of Object.entries(tree)) {
        const path = prefix ? `${prefix}/${name}` : name;
        if ("file" in node) {
            result[path] = String(node.file.contents);
        } else if ("directory" in node) {
            Object.assign(result, flattenTreeToFiles(node.directory, path));
        }
    }
    return result;
}

function getMonacoLanguage(path: string): string {
    const ext = path.split(".").pop()?.toLowerCase();
    switch (ext) {
        case "jsx":
        case "js":
            return "javascript";
        case "tsx":
        case "ts":
            return "typescript";
        case "json":
            return "json";
        case "html":
            return "html";
        case "css":
            return "css";
        case "md":
            return "markdown";
        default:
            return "plaintext";
    }
}

const INITIAL_FILES = flattenTreeToFiles(FRONTEND_QUESTION_PROJECT);
const INITIAL_FILE_PATHS = Object.keys(INITIAL_FILES).sort();
const DEFAULT_OPEN_FILE =
    INITIAL_FILE_PATHS.find((path) => path === "src/App.jsx") ??
    INITIAL_FILE_PATHS[0] ??
    "";

export interface FrontendQuestion {
    title: string;
    description: string;
    requirements?: string[];
}

const DEFAULT_QUESTION: FrontendQuestion = {
    title: "Build a Counter Component",
    description:
        "Implement a simple Counter component in React. The user should be able to increment, decrement, and reset the count. Display the current value clearly.",
    requirements: [
        "Use the useState hook",
        "Provide Increment, Decrement, and Reset controls",
        "Prevent the count from going below 0",
        "Show the current count prominently",
    ],
};

interface FrontendWebContainerProps {
    question?: FrontendQuestion;
}

export function FrontendWebContainer({
    question = DEFAULT_QUESTION,
}: FrontendWebContainerProps) {
    const webContainerRef = useRef<WebContainer | null>(null);
    const installProcessRef = useRef<{ kill: () => void } | null>(null);
    const devProcessRef = useRef<{ kill: () => void } | null>(null);
    const logsEndRef = useRef<HTMLDivElement | null>(null);
    const writeTimerRef = useRef<number | undefined>(undefined);
    const isMountedRef = useRef(false);

    const [files, setFiles] = useState<Record<string, string>>(INITIAL_FILES);
    const [currentFile, setCurrentFile] = useState<string>(DEFAULT_OPEN_FILE);
    const [isBooting, setIsBooting] = useState(false);
    const [previewUrl, setPreviewUrl] = useState("");
    const [logs, setLogs] = useState<string[]>([]);
    const [errorMessage, setErrorMessage] = useState("");
    const [isTerminalOpen, setIsTerminalOpen] = useState(true);

    const filePaths = useMemo(() => Object.keys(files).sort(), [files]);

    const appendLog = useCallback((msg: string) => {
        const clean = sanitizeOutput(msg);
        if (!clean) return;
        setLogs((prev) => [...prev, clean]);
    }, []);

    const startWebContainer = useCallback(async () => {
        if (webContainerRef.current) return;

        setIsBooting(true);
        setErrorMessage("");
        appendLog("[boot] starting WebContainer\n");

        try {
            const { WebContainer } = await import("@webcontainer/api");
            const wc = await WebContainer.boot();
            webContainerRef.current = wc;

            wc.on("server-ready", (_port, url) => {
                setPreviewUrl(url);
                appendLog(`[server] ready at ${url}\n`);
            });

            await wc.mount(FRONTEND_QUESTION_PROJECT);
            appendLog("[mount] files mounted\n");
            isMountedRef.current = true;

            const installArgs = ["install", "--no-progress"];
            appendLog(`$ npm ${installArgs.join(" ")}\n`);
            const install = await wc.spawn("npm", installArgs);
            installProcessRef.current = install;
            install.output.pipeTo(
                new WritableStream({ write: (chunk) => appendLog(String(chunk)) }),
            );
            const installCode = await install.exit;
            if (installCode !== 0) {
                throw new Error(`npm install failed with code ${installCode}`);
            }
            appendLog("[install] done\n");

            const devArgs = ["run", "dev"];
            appendLog(`$ npm ${devArgs.join(" ")}\n`);
            const dev = await wc.spawn("npm", devArgs);
            devProcessRef.current = dev;
            dev.output.pipeTo(
                new WritableStream({ write: (chunk) => appendLog(String(chunk)) }),
            );
        } catch (err) {
            const msg =
                err instanceof Error ? err.message : "Unable to start WebContainer";
            setErrorMessage(msg);
            appendLog(`[error] ${msg}\n`);
        } finally {
            setIsBooting(false);
        }
    }, [appendLog]);

    useEffect(() => {
        const t = window.setTimeout(() => void startWebContainer(), 0);
        return () => {
            window.clearTimeout(t);
            if (writeTimerRef.current) window.clearTimeout(writeTimerRef.current);
            installProcessRef.current?.kill();
            devProcessRef.current?.kill();
            webContainerRef.current?.teardown();
            webContainerRef.current = null;
            isMountedRef.current = false;
        };
    }, [startWebContainer]);

    useEffect(() => {
        if (!isTerminalOpen) return;
        logsEndRef.current?.scrollIntoView({ behavior: "smooth", block: "end" });
    }, [logs, isTerminalOpen]);

    const handleEditorChange = useCallback(
        (value: string | undefined) => {
            if (value === undefined) return;
            setFiles((prev) => ({ ...prev, [currentFile]: value }));
            if (writeTimerRef.current) {
                window.clearTimeout(writeTimerRef.current);
            }
            writeTimerRef.current = window.setTimeout(() => {
                const wc = webContainerRef.current;
                if (wc && isMountedRef.current) {
                    void wc.fs.writeFile(currentFile, value);
                }
            }, 200);
        },
        [currentFile],
    );

    return (
        <div className="flex h-[calc(100dvh-49px)] flex-col bg-[#0e0e0e] text-zinc-200">
            <header className="flex items-center justify-between border-b border-white/8 bg-[#0a0a0a] px-4 py-2">
                <h1 className="text-sm font-semibold text-white">
                    Frontend Sandbox
                </h1>
                <div className="flex items-center gap-2">
                    {errorMessage ? (
                        <span className="rounded-md border border-red-500/30 bg-red-500/10 px-2 py-1 text-xs text-red-200">
                            {errorMessage}
                        </span>
                    ) : null}
                    <button
                        type="button"
                        disabled={isBooting}
                        onClick={() => {
                            if (!isBooting) void startWebContainer();
                        }}
                        className="rounded-md border border-white/15 bg-white/5 px-3 py-1.5 text-xs text-zinc-100 transition-colors hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-60"
                    >
                        {isBooting ? "Starting..." : "Restart"}
                    </button>
                </div>
            </header>

            <div className="flex min-h-0 flex-1">
                <aside className="w-[320px] shrink-0 overflow-y-auto border-r border-white/8 bg-[#101010] px-5 py-4">
                    <h2 className="mb-3 text-base font-semibold text-white">
                        {question.title}
                    </h2>
                    <p className="mb-5 text-sm leading-relaxed text-zinc-300">
                        {question.description}
                    </p>
                    {question.requirements && question.requirements.length > 0 ? (
                        <>
                            <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-widest text-zinc-500">
                                Requirements
                            </h3>
                            <ul className="list-disc space-y-1.5 pl-5 text-sm text-zinc-300">
                                {question.requirements.map((requirement) => (
                                    <li key={requirement}>{requirement}</li>
                                ))}
                            </ul>
                        </>
                    ) : null}
                </aside>

                <section className="flex min-w-0 flex-1 flex-col border-r border-white/8 bg-[#1e1e1e]">
                    <div className="flex shrink-0 overflow-x-auto border-b border-white/8 bg-[#0a0a0a]">
                        {filePaths.map((path) => {
                            const isActive = path === currentFile;
                            return (
                                <button
                                    key={path}
                                    type="button"
                                    onClick={() => setCurrentFile(path)}
                                    className={`shrink-0 border-r border-white/8 px-3 py-2 font-mono text-xs whitespace-nowrap transition-colors ${
                                        isActive
                                            ? "bg-[#1e1e1e] text-white"
                                            : "text-zinc-400 hover:bg-white/5 hover:text-zinc-100"
                                    }`}
                                >
                                    {path}
                                </button>
                            );
                        })}
                    </div>
                    <div className="min-h-0 flex-1">
                        {currentFile ? (
                            <MonacoEditor
                                height="100%"
                                path={currentFile}
                                language={getMonacoLanguage(currentFile)}
                                value={files[currentFile] ?? ""}
                                onChange={handleEditorChange}
                                theme="vs-dark"
                                options={{
                                    fontSize: 13,
                                    fontFamily:
                                        "'JetBrains Mono', 'Fira Code', monospace",
                                    fontLigatures: true,
                                    minimap: { enabled: false },
                                    scrollBeyondLastLine: false,
                                    automaticLayout: true,
                                    tabSize: 2,
                                    padding: { top: 12, bottom: 12 },
                                    renderLineHighlight: "line",
                                    overviewRulerLanes: 0,
                                    hideCursorInOverviewRuler: true,
                                    scrollbar: {
                                        verticalScrollbarSize: 6,
                                        horizontalScrollbarSize: 6,
                                    },
                                }}
                            />
                        ) : (
                            <div className="flex h-full items-center justify-center text-sm text-zinc-400">
                                No file open
                            </div>
                        )}
                    </div>
                </section>

                <section className="flex min-w-0 flex-1 flex-col bg-[#141414]">
                    {previewUrl ? (
                        <iframe
                            title="Frontend Question Preview"
                            src={previewUrl}
                            className="h-full w-full flex-1"
                            allow="clipboard-write; cross-origin-isolated; fullscreen"
                        />
                    ) : (
                        <div className="flex flex-1 items-center justify-center text-sm text-zinc-400">
                            Booting WebContainer preview...
                        </div>
                    )}
                </section>
            </div>

            <section
                className={`flex shrink-0 flex-col border-t border-white/8 bg-[#0a0a0a] transition-[height] duration-150 ${
                    isTerminalOpen ? "h-64" : "h-9"
                }`}
            >
                <button
                    type="button"
                    onClick={() => setIsTerminalOpen((v) => !v)}
                    className="flex h-9 shrink-0 items-center justify-between border-b border-white/8 px-4 text-[11px] font-semibold uppercase tracking-widest text-zinc-400 hover:text-zinc-100"
                    aria-expanded={isTerminalOpen}
                >
                    <span>Terminal</span>
                    <span className="text-zinc-500">
                        {isTerminalOpen ? "Hide" : "Show"}
                    </span>
                </button>
                {isTerminalOpen ? (
                    <div className="flex-1 overflow-auto bg-black/40 px-4 py-3 font-mono text-xs whitespace-pre-wrap text-zinc-300">
                        {logs.length > 0
                            ? logs.join("")
                            : "WebContainer logs will appear here..."}
                        <div ref={logsEndRef} />
                    </div>
                ) : null}
            </section>
        </div>
    );
}
