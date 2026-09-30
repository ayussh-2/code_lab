import type { KeyboardEvent } from "react";

interface ResizeHandleProps {
  width: number;
  onKeyDown: (event: KeyboardEvent<HTMLDivElement>) => void;
  onPointerDown: () => void;
}

export function ResizeHandle({
  width,
  onKeyDown,
  onPointerDown,
}: ResizeHandleProps) {
  return (
    <div
      role="separator"
      aria-label="Resize left panel"
      aria-valuemin={18}
      aria-valuemax={50}
      aria-valuenow={Math.round(width)}
      tabIndex={0}
      onKeyDown={onKeyDown}
      onPointerDown={onPointerDown}
      className="w-1 shrink-0 cursor-col-resize bg-zinc-800 transition-colors hover:bg-blue-500 focus:bg-blue-500 focus:outline-none"
    />
  );
}
