import Link from "next/link";
import { Sparkles } from "lucide-react";
import type { AccountSummary } from "@/types";

// Indicador NO intrusivo del cupo del plan Free. Solo aparece cuando el usuario
// se acerca o llega al límite, para no molestar al resto. Pro no tiene límite → no se
// muestra.
const NEAR_THRESHOLD = 0.8;

export function PlanQuotaBanner({ summary }: { summary: AccountSummary | null }) {
  if (!summary || summary.plan !== "free") return null;
  const limit = summary.activeUrlsLimit;
  if (typeof limit !== "number" || limit <= 0) return null;

  const used = summary.activeUrlsCount;
  if (used / limit < NEAR_THRESHOLD) return null;

  const atLimit = used >= limit;

  return (
    <div className="flex flex-wrap items-center gap-x-2 gap-y-1 rounded-lg border border-amber-200 bg-amber-50 px-4 py-2.5 text-sm text-amber-800 dark:border-amber-900/50 dark:bg-amber-900/20 dark:text-amber-300">
      <Sparkles className="size-4 shrink-0" aria-hidden />
      <span>
        {atLimit
          ? `Llegaste al límite: ${used} de ${limit} URLs activas.`
          : `${used} de ${limit} URLs activas en tu plan Free.`}
      </span>
      <Link
        href="/dashboard/account"
        className="font-semibold underline underline-offset-2 hover:text-amber-900 dark:hover:text-amber-200"
      >
        Mejora a Pro
      </Link>
    </div>
  );
}
