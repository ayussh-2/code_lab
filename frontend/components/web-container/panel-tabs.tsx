import type { ActivePanel } from "./types";

interface PanelTabsProps {
  activePanel: ActivePanel;
  onChange: (panel: ActivePanel) => void;
}

export function PanelTabs({ activePanel, onChange }: PanelTabsProps) {
  return (
    <div
      className="flex rounded-md border border-zinc-700 bg-zinc-950 p-0.5"
      role="tablist"
      aria-label="Left panel"
    >
      {(["question", "files"] as const).map((panel) => (
        <button
          key={panel}
          type="button"
          role="tab"
          aria-selected={activePanel === panel}
          onClick={() => onChange(panel)}
          className={`rounded px-3 py-1.5 text-xs font-medium capitalize ${
            activePanel === panel
              ? "bg-zinc-100 text-zinc-900"
              : "text-zinc-400 hover:text-zinc-100"
          }`}
        >
          {panel}
        </button>
      ))}
    </div>
  );
}
