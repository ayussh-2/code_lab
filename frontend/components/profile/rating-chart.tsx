interface RatingChartProps {
  points: Array<{ rating: number; recorded_at: string }>;
}

export function RatingChart({ points }: RatingChartProps) {
  if (points.length === 0) {
    return (
      <p className="text-sm text-zinc-500">No rating history yet.</p>
    );
  }

  const width = 640;
  const height = 160;
  const padX = 8;
  const padY = 12;
  const ratings = points.map((p) => p.rating);
  const minR = Math.min(...ratings) - 20;
  const maxR = Math.max(...ratings) + 20;
  const span = Math.max(maxR - minR, 1);

  const coords = points.map((p, i) => {
    const x =
      padX +
      (i / Math.max(points.length - 1, 1)) * (width - padX * 2);
    const y =
      height -
      padY -
      ((p.rating - minR) / span) * (height - padY * 2);
    return { x, y };
  });

  const polyline = coords.map((c) => `${c.x},${c.y}`).join(" ");
  const latest = points[points.length - 1]?.rating ?? 0;

  return (
    <div className="space-y-2">
      <div className="flex items-baseline justify-between">
        <p className="text-[11px] uppercase tracking-widest text-zinc-500">
          Rating
        </p>
        <p className="font-mono text-lg font-semibold text-white">{latest}</p>
      </div>
      <svg
        viewBox={`0 0 ${width} ${height}`}
        className="h-40 w-full text-[#1cbf73]"
        role="img"
        aria-label="Rating history chart"
      >
        <polyline
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          points={polyline}
        />
        {coords.map((c, i) => (
          <circle
            key={`${c.x}-${i}`}
            cx={c.x}
            cy={c.y}
            r={3}
            fill="currentColor"
          />
        ))}
      </svg>
    </div>
  );
}
