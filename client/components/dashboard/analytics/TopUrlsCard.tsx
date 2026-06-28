import Link from "next/link";
import { Trophy } from "lucide-react";
import { Card } from "@/components/ui";
import type { ShortURL } from "@/types";

// Ranking de URLs con más clics. El endpoint agregado GetStats NO desglosa por
// URL, así que el ranking usa el contador totalClicks por URL que ya expone el dominio
// (real, no inventado). IMPORTANTE: es HISTÓRICO (all-time), no acotado al rango — se
// etiqueta como tal para no implicar datos por-rango que el backend no entrega. Se nutre
// de la página actual del store de URLs (hasta 20); para la mayoría de cuentas es exacto.
const TOP_N = 5;

function formatNumber(n: number): string {
  return new Intl.NumberFormat("es").format(n);
}

export function TopUrlsCard({ urls }: { urls: ShortURL[] }) {
  const ranked = [...urls]
    .filter((u) => u.totalClicks > 0)
    .sort((a, b) => b.totalClicks - a.totalClicks)
    .slice(0, TOP_N);

  if (ranked.length === 0) return null;

  const max = ranked[0].totalClicks;

  return (
    <Card>
      <div className="mb-3 flex items-center gap-2">
        <span className="text-brand-600 dark:text-brand-300" aria-hidden>
          <Trophy className="size-4" />
        </span>
        <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
          URLs con más clics
        </h2>
        <span className="ml-auto text-xs text-zinc-400">Histórico</span>
      </div>

      <ol className="space-y-3">
        {ranked.map((url, i) => {
          const pct = max > 0 ? Math.round((url.totalClicks / max) * 100) : 0;
          return (
            <li key={url.id} className="flex items-center gap-3">
              <span className="w-4 shrink-0 text-sm font-semibold tabular-nums text-zinc-400">
                {i + 1}
              </span>
              <div className="min-w-0 flex-1">
                <div className="mb-1 flex items-center justify-between gap-2 text-sm">
                  <Link
                    href={`/dashboard/urls/${url.shortCode}`}
                    className="truncate font-medium text-brand-600 hover:underline dark:text-brand-400"
                  >
                    /{url.shortCode}
                  </Link>
                  <span className="shrink-0 tabular-nums text-zinc-500 dark:text-zinc-400">
                    {formatNumber(url.totalClicks)}
                  </span>
                </div>
                <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800">
                  <div
                    className="h-full rounded-full bg-brand-500"
                    style={{ width: `${pct}%` }}
                  />
                </div>
              </div>
            </li>
          );
        })}
      </ol>
    </Card>
  );
}
