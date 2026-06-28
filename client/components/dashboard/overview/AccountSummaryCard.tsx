import { BarChart3, CalendarDays, Link2, Radio } from "lucide-react";
import { Card, Spinner } from "@/components/ui";
import type { AccountSummary } from "@/types";
import { StatCard } from "./StatCard";

function formatNumber(n: number): string {
  return new Intl.NumberFormat("es").format(n);
}

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("es", {
      day: "numeric",
      month: "short",
      year: "numeric",
    }).format(new Date(iso));
  } catch {
    return "—";
  }
}

// Barra de progreso del cupo de URLs activas (solo plan Free, que tiene límite).
function QuotaBar({ used, limit }: { used: number; limit: number }) {
  const pct = limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0;
  const near = pct >= 80;
  return (
    <div
      className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800"
      role="progressbar"
      aria-valuenow={used}
      aria-valuemin={0}
      aria-valuemax={limit}
      aria-label="Cupo de URLs activas"
    >
      <div
        className={near ? "h-full bg-amber-500" : "h-full bg-brand-600"}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

export function AccountSummaryCard({
  summary,
  liveClicks,
  isLoading,
}: {
  summary: AccountSummary | null;
  liveClicks: number;
  isLoading: boolean;
}) {
  if (isLoading && !summary) {
    return (
      <Card className="flex h-32 items-center justify-center">
        <Spinner size={24} className="text-brand-600 dark:text-brand-400" />
      </Card>
    );
  }

  if (!summary) {
    return (
      <Card className="text-sm text-zinc-500 dark:text-zinc-400">
        No se pudo cargar el resumen de tu cuenta.
      </Card>
    );
  }

  const isFree = summary.plan === "free";
  const limit = summary.activeUrlsLimit ?? null;
  const showQuota = isFree && typeof limit === "number";

  return (
    <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <StatCard
        label="Clics totales"
        value={formatNumber(summary.totalClicks)}
        icon={<BarChart3 className="size-4" />}
        hint="Acumulado de todas tus URLs"
      />

      <StatCard
        label="URLs activas"
        value={formatNumber(summary.activeUrlsCount)}
        icon={<Link2 className="size-4" />}
        hint={
          showQuota
            ? `de ${formatNumber(limit as number)} en plan Free`
            : "Ilimitadas en plan Pro"
        }
      >
        {showQuota && (
          <QuotaBar used={summary.activeUrlsCount} limit={limit as number} />
        )}
      </StatCard>

      <StatCard
        label="Clics en vivo"
        value={`+${formatNumber(liveClicks)}`}
        icon={<Radio className="size-4" />}
        hint="Desde que abriste esta página"
        highlight
      />

      <StatCard
        label="Miembro desde"
        value={formatDate(summary.createdAt)}
        icon={<CalendarDays className="size-4" />}
        hint={summary.email}
      />
    </div>
  );
}
