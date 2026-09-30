import type { KeyboardEvent } from "react";

interface ResizeHandleProps {
  width: number;
  label?: string;
  minWidth?: number;
  maxWidth?: number;
  onKeyDown: (event: KeyboardEvent<HTMLDivElement>) => void;
  onPointerDown: () => void;
}

export function ResizeHandle({
  width,
  label = "Resize left panel",
  minWidth = 18,
  maxWidth = 50,
  onKeyDown,
  onPointerDown,
}: ResizeHandleProps) {
  return (
    <div
      role="separator"
      aria-label={label}
      aria-valuemin={minWidth}
      aria-valuemax={maxWidth}
      aria-valuenow={Math.round(width)}
      tabIndex={0}
      onKeyDown={onKeyDown}
      onPointerDown={onPointerDown}
      className="w-1 shrink-0 cursor-col-resize bg-zinc-800 transition-colors hover:bg-blue-500 focus:bg-blue-500 focus:outline-none"
    />
  );
}
