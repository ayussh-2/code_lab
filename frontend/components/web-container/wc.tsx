"use client";

import { Button } from "../ui/button";
import { CodeEditorPanel } from "./code-editor-panel";
import { FilesPanel, createTreeElements } from "./files-panel";
import { PanelTabs } from "./panel-tabs";
import { PreviewPanel } from "./preview-panel";
import { QuestionPanel } from "./question-panel";
import { ResizeHandle } from "./resize-handle";
import { TerminalPanel } from "./terminal-panel";
import type { ActivePanel, WebContProps } from "./types";
import { useWebContainer } from "@/hooks/use-webcontainer";
import { useEffect, useMemo, useRef, useState } from "react";

export default function WebCont({ files, question }: WebContProps) {
  const [filesPanelWidth, setFilesPanelWidth] = useState(28);
  const [isResizing, setIsResizing] = useState(false);
  const [previewPanelWidth, setPreviewPanelWidth] = useState(38);
  const [isResizingPreview, setIsResizingPreview] = useState(false);
  const [activePanel, setActivePanel] = useState<ActivePanel>("question");
  const [isPanelCollapsed, setIsPanelCollapsed] = useState(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const workspaceRef = useRef<HTMLDivElement>(null);
  const tree = useMemo(() => createTreeElements(files), [files]);
  const [selectedFile, setSelectedFile] = useState<string | undefined>(
    tree.firstFileId,
  );
  const {
    status,
    error,
    terminalOutput,
    readFile,
    writeFile,
  } = useWebContainer({
    files,
    iframeRef,
  });

  useEffect(() => {
    if (!isResizing) return;

    function handlePointerMove(event: PointerEvent) {
      const workspace = workspaceRef.current;
      if (!workspace) return;

      const { left, width } = workspace.getBoundingClientRect();
      const nextWidth = ((event.clientX - left) / width) * 100;
      setFilesPanelWidth(Math.min(50, Math.max(18, nextWidth)));
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

  useEffect(() => {
    if (!isResizingPreview) return;

    function handlePointerMove(event: PointerEvent) {
      const workspace = workspaceRef.current;
      if (!workspace) return;

      const { right, width } = workspace.getBoundingClientRect();
      const nextWidth = ((right - event.clientX) / width) * 100;
      setPreviewPanelWidth(Math.min(60, Math.max(25, nextWidth)));
    }

    function stopResizing() {
      setIsResizingPreview(false);
    }

    window.addEventListener("pointermove", handlePointerMove);
    window.addEventListener("pointerup", stopResizing);
    return () => {
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", stopResizing);
    };
  }, [isResizingPreview]);

  function handleResizeKeyDown(event: React.KeyboardEvent<HTMLDivElement>) {
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      setFilesPanelWidth((width) => Math.max(18, width - 2));
    }
    if (event.key === "ArrowRight") {
      event.preventDefault();
      setFilesPanelWidth((width) => Math.min(50, width + 2));
    }
  }

  function handlePreviewResizeKeyDown(
    event: React.KeyboardEvent<HTMLDivElement>,
  ) {
    if (event.key === "ArrowLeft") {
      event.preventDefault();
      setPreviewPanelWidth((width) => Math.min(60, width + 2));
    }
    if (event.key === "ArrowRight") {
      event.preventDefault();
      setPreviewPanelWidth((width) => Math.max(25, width - 2));
    }
  }

  console.log(status);

  return (
    <div
      ref={workspaceRef}
      className={`relative flex h-full min-h-0 w-full overflow-hidden bg-zinc-950 text-zinc-200 ${
        isResizing || isResizingPreview ? "select-none" : ""
      }`}
    >
      <aside
        className="flex h-full min-w-0 shrink-0 flex-col overflow-hidden bg-zinc-900 transition-[width] duration-200"
        style={{ width: isPanelCollapsed ? 0 : `${filesPanelWidth}%` }}
      >
        <div className="flex shrink-0 items-center justify-between border-b border-zinc-800 px-3 py-2">
          <PanelTabs activePanel={activePanel} onChange={setActivePanel} />
          <Button
            type="button"
            aria-label="Collapse left panel"
            title="Collapse left panel"
            onClick={() => setIsPanelCollapsed(true)}
          >
            <span aria-hidden="true">‹</span>
          </Button>
        </div>

        {activePanel === "question" ? (
          <QuestionPanel question={question} />
        ) : (
          <FilesPanel
            files={tree}
            onSelectFile={(path) => setSelectedFile(path)}
          />
        )}
      </aside>

      {!isPanelCollapsed && (
        <ResizeHandle
          width={filesPanelWidth}
          onKeyDown={handleResizeKeyDown}
          onPointerDown={() => setIsResizing(true)}
        />
      )}

      {isPanelCollapsed && (
        <button
          type="button"
          aria-label="Expand left panel"
          title="Expand left panel"
          onClick={() => setIsPanelCollapsed(false)}
          className="absolute left-2 top-2 z-20 rounded border border-zinc-700 bg-zinc-900 px-2 py-1 text-zinc-300 shadow-lg hover:bg-zinc-800"
        >
          <span aria-hidden="true">›</span>
        </button>
      )}

      <CodeEditorPanel
        filePath={selectedFile}
        readFile={readFile}
        writeFile={writeFile}
      />
      <ResizeHandle
        width={previewPanelWidth}
        label="Resize preview panel"
        minWidth={25}
        maxWidth={60}
        onKeyDown={handlePreviewResizeKeyDown}
        onPointerDown={() => setIsResizingPreview(true)}
      />
      <div
        className="flex h-full min-w-70"
        style={{ width: `${previewPanelWidth}%` }}
      >
        <PreviewPanel iframeRef={iframeRef} status={status} error={error} />
      </div>
      <TerminalPanel output={terminalOutput} />
    </div>
  );
}
