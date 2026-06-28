"use client";

import { Lock } from "lucide-react";
import { cn } from "@/lib/utils/cn";
import type { Plan, TimeRange } from "@/types";


const OPTIONS: { value: TimeRange; label: string; proOnly: boolean }[] = [
  { value: "24h", label: "24h", proOnly: false },
  { value: "7d", label: "7 días", proOnly: true },
  { value: "30d", label: "30 días", proOnly: true },
];

export function RangeSelector({
  value,
  onChange,
  plan,
}: {
  value: TimeRange;
  onChange: (range: TimeRange) => void;
  plan: Plan | null;
}) {
  const isFree = plan === "free";

  return (
    <div
      role="group"
      aria-label="Rango de tiempo"
      className="inline-flex items-center gap-1 rounded-lg border border-zinc-200 bg-zinc-50 p-1 dark:border-zinc-800 dark:bg-zinc-900"
    >
      {OPTIONS.map((opt) => {
        const locked = isFree && opt.proOnly;
        const active = value === opt.value;
        return (
          <button
            key={opt.value}
            type="button"
            onClick={() => !locked && onChange(opt.value)}
            disabled={locked}
            aria-pressed={active}
            title={locked ? "Requiere plan Pro" : undefined}
            className={cn(
              "inline-flex items-center gap-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
              active
                ? "bg-white text-brand-700 shadow-sm dark:bg-zinc-800 dark:text-brand-300"
                : "text-zinc-600 hover:text-zinc-900 dark:text-zinc-400 dark:hover:text-zinc-100",
              locked && "cursor-not-allowed opacity-50 hover:text-zinc-600",
            )}
          >
            {opt.label}
            {locked && <Lock className="size-3" aria-hidden />}
          </button>
        );
      })}
    </div>
  );
}
