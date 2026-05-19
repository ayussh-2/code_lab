import { useEffect, useRef } from "react";
import type { ActivityDay } from "@/lib/profile";

interface ActivityHeatmapProps {
  days: ActivityDay[];
}

function formatDateKey(date: Date): string {
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return `${y}-${m}-${d}`;
}

function buildWeekGrid(days: ActivityDay[]) {
  const byDate = new Map(days.map((d) => [d.date, d.count]));
  const end = new Date();
  end.setHours(0, 0, 0, 0);
  const start = new Date(end);
  start.setDate(start.getDate() - 364);
  start.setDate(start.getDate() - start.getDay());

  const weeks: Array<Array<{ date: string; count: number }>> = [];
  let week: Array<{ date: string; count: number }> = [];
  const cursor = new Date(start);

  while (cursor <= end) {
    const key = formatDateKey(cursor);
    week.push({ date: key, count: byDate.get(key) ?? 0 });
    if (cursor.getDay() === 6) {
      weeks.push(week);
      week = [];
    }
    cursor.setDate(cursor.getDate() + 1);
  }
  if (week.length > 0) {
    weeks.push(week);
  }
  return weeks;
}

function level(count: number): number {
  if (count === 0) return 0;
  if (count <= 2) return 1;
  if (count <= 5) return 2;
  if (count <= 10) return 3;
  return 4;
}

const LEVEL_CLASS = [
  "bg-white/[0.04]",
  "bg-[#1cbf73]/20",
  "bg-[#1cbf73]/40",
  "bg-[#1cbf73]/65",
  "bg-[#1cbf73]",
];

export function ActivityHeatmap({ days }: ActivityHeatmapProps) {
  const scrollRef = useRef<HTMLDivElement>(null);
  const weeks = buildWeekGrid(days);
  const max = Math.max(
    ...weeks.flat().map((c) => c.count),
    0,
  );

  useEffect(() => {
    const el = scrollRef.current;
    if (el) {
      el.scrollLeft = el.scrollWidth;
    }
  }, [days]);

  return (
    <div className="space-y-3">
      <p className="text-[11px] uppercase tracking-widest text-zinc-500">
        Activity (last year)
      </p>
      <div
        ref={scrollRef}
        className="overflow-x-auto pb-1"
      >
        <div className="inline-flex gap-1">
          {weeks.map((week, weekIndex) => (
            <div key={weekIndex} className="flex flex-col gap-1">
              {week.map((cell) => (
                <div
                  key={cell.date}
                  title={`${cell.date}: ${cell.count} submission${cell.count === 1 ? "" : "s"}`}
                  className={`h-3 w-3 shrink-0 rounded-sm ${LEVEL_CLASS[level(cell.count)]}`}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
      <p className="text-xs text-zinc-500">
        {max > 0
          ? `Max ${max} submissions in a day`
          : "No submissions in the last year"}
      </p>
    </div>
  );
}
