"use client";

import { useEffect, useState } from "react";

interface TerminalPanelProps {
  output: string;
}

export function TerminalPanel({ output }: TerminalPanelProps) {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const [panelHeight, setPanelHeight] = useState(360);
  const [isResizing, setIsResizing] = useState(false);

  useEffect(() => {
    if (!isResizing) return;

    function handlePointerMove(event: PointerEvent) {
      const nextHeight = window.innerHeight - event.clientY;
      setPanelHeight(Math.min(window.innerHeight, Math.max(120, nextHeight)));
    }

    function stopResizing() {
      setIsResizing(false);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", stopResizing);
    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", stopResizing);
    };
  }, [isResizing]);

  function handleResizeKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === "ArrowUp") {
      event.preventDefault();
      setPanelHeight((height) => Math.min(window.innerHeight, height + 40));
    }
    if (event.key === "ArrowDown") {
      event.preventDefault();
      setPanelHeight((height) => Math.max(120, height - 40));
    }
  }

  return (
    <section
      className="fixed inset-x-0 bottom-0 z-100 flex flex-col border-t border-zinc-700 bg-zinc-950 text-zinc-100 shadow-2xl"
      style={{ height: isCollapsed ? 32 : panelHeight }}
    >
      {!isCollapsed && (
        <div
          role="separator"
          aria-label="Resize terminal"
          aria-valuemin={120}
          aria-valuemax={typeof window === "undefined" ? 0 : window.innerHeight}
          aria-valuenow={panelHeight}
          tabIndex={0}
          onKeyDown={handleResizeKeyDown}
          onPointerDown={() => setIsResizing(true)}
          className="h-1 shrink-0 cursor-row-resize bg-zinc-800 transition-colors hover:bg-blue-500 focus:bg-blue-500 focus:outline-none"
        />
      )}

      {!isCollapsed && (
        <>
          <pre className="min-h-0 flex-1 overflow-auto whitespace-pre-wrap wrap-break-word p-4 font-mono text-xs leading-5 text-zinc-300">
            {output || "Waiting for WebContainer output..."}
          </pre>
          <div className="flex shrink-0 justify-end border-t border-zinc-800 bg-zinc-900 p-2">
            <button
              type="button"
              onClick={() => setIsCollapsed(true)}
              aria-label="Collapse terminal"
              className="rounded bg-zinc-800 px-3 py-2 text-xs font-semibold text-zinc-300 hover:bg-zinc-700"
            >
              Collapse
            </button>
          </div>
        </>
      )}
      {isCollapsed && (
        <button
          type="button"
          onClick={() => setIsCollapsed(false)}
          aria-label="Expand terminal"
          className="h-8 w-full text-xs font-semibold text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100"
        >
          Expand terminal
        </button>
      )}
    </section>
  );
}
