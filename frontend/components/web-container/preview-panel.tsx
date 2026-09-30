import type { RefObject } from "react";

interface PreviewPanelProps {
  iframeRef: RefObject<HTMLIFrameElement | null>;
  status: string;
  error: string | null;
}

export function PreviewPanel({ iframeRef, status, error }: PreviewPanelProps) {
  const statusText =
    status === "booting"
      ? "Starting WebContainer..."
      : status === "mounting"
        ? "Loading project..."
        : status === "starting"
          ? "Starting development server..."
          : status === "ready"
            ? "Ready"
            : "Loading preview...";

  return (
    <section className="relative min-w-0 flex-1 bg-white">
      {status !== "ready" && !error && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-white">
          <div className="flex flex-col items-center gap-4">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-gray-200 border-t-black" />
            <p className="text-sm text-gray-600">{statusText}</p>
          </div>
        </div>
      )}
      {error && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-white">
          <div className="max-w-md px-6 text-center">
            <p className="font-semibold text-red-600">
              Failed to start WebContainer
            </p>
            <p className="mt-2 text-sm text-gray-600">{error}</p>
          </div>
        </div>
      )}
      <iframe
        ref={iframeRef}
        title="WebContainer preview"
        className="block h-full min-h-80 w-full border-0"
      />
    </section>
  );
}
