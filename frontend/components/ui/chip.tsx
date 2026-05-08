import type { ReactNode } from "react";

interface ChipProps {
  children: ReactNode;
  className?: string;
}

export function Chip({ children, className = "" }: ChipProps) {
  return (
    <span
      className={`rounded-full border border-[#2a2a2a] bg-[#1d1d1d] px-2 py-0.5 text-xs font-medium text-zinc-300 ${className}`.trim()}
    >
      {children}
    </span>
  );
}
