import type { ReactNode } from "react";
import { cn } from "@/lib/utils/cn";
import { Card } from "@/components/ui";

export function StatCard({
  label,
  value,
  icon,
  hint,
  highlight,
  children,
}: {
  label: string;
  value: ReactNode;
  icon: ReactNode;
  hint?: ReactNode;
  /** Resalta la tarjeta con el acento de marca (ej. la de clics en vivo). */
  highlight?: boolean;
  /** Contenido extra opcional (ej. una barra de progreso de cupo). */
  children?: ReactNode;
}) {
  return (
    <Card className="flex flex-col gap-3 p-5">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-zinc-500 dark:text-zinc-400">
          {label}
        </span>
        <span
          className={cn(
            "flex size-8 items-center justify-center rounded-lg",
            highlight
              ? "bg-accent-100 text-accent-700 dark:bg-accent-900/40 dark:text-accent-300"
              : "bg-brand-50 text-brand-600 dark:bg-brand-900/40 dark:text-brand-300",
          )}
          aria-hidden
        >
          {icon}
        </span>
      </div>
      <div>
        <p className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          {value}
        </p>
        {hint && (
          <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">{hint}</p>
        )}
      </div>
      {children}
    </Card>
  );
}
