import type { ReactNode } from "react";
import { scaleOrdinal, schemeTableau10 } from "d3";
import { Card } from "@/components/ui";

// Tarjeta de desglose genérica. Una sola
// implementación reutilizada para país, dispositivo y navegador (en vez de tres
// componentes casi idénticos), siguiendo la regla de reutilización del CLAUDE.md. Lista
// con barras de progreso simples — más legible y liviana que un gráfico D3 para datos
// categóricos pequeños. La usan tanto Detalle de URL como Analytics vía StatsPanels.

const TOP_N = 8;

function formatNumber(n: number): string {
  return new Intl.NumberFormat("es").format(n);
}

type Row = { key: string; label: string; value: number; color: string };

export function BreakdownCard({
  title,
  icon,
  data,
  emptyLabel = "Sin datos todavía.",
  labelMap,
}: {
  title: string;
  icon: ReactNode;
  data: Record<string, number>;
  emptyLabel?: string;
  /** Traducción/legibilidad opcional de las claves (ej. "ES" → "España"). */
  labelMap?: (key: string) => string;
}) {
  const sorted = Object.entries(data)
    .filter(([, v]) => v > 0)
    .sort((a, b) => b[1] - a[1]);

  // Solo agrupamos en "Otros" si realmente ahorra más de una fila: con TOP_N+1 entradas
  // se muestran todas (evita un poco natural "Otros (1)").
  const grouped = sorted.length > TOP_N + 1;
  const individual = grouped ? sorted.slice(0, TOP_N) : sorted;
  const rest = grouped ? sorted.slice(TOP_N) : [];
  const restTotal = rest.reduce((sum, [, v]) => sum + v, 0);

  // Escala categórica de D3: colores pensados para distinguir muchas categorías a la vez,
  // en vez de elegir hex a mano. Determinista por clave (orden del dominio).
  const color = scaleOrdinal<string, string>()
    .domain(individual.map(([key]) => key))
    .range(schemeTableau10 as string[]);

  const rows: Row[] = individual.map(([key, value]) => ({
    key,
    label: labelMap ? labelMap(key) : key,
    value,
    color: color(key),
  }));

  if (grouped && restTotal > 0) {
    rows.push({
      key: "__otros__",
      label: `Otros (${rest.length})`,
      value: restTotal,
      color: "#a1a1aa", // zinc-400: neutro, señala que es un agregado, no una categoría
    });
  }

  // Ancho de barra relativo al mayor valor mostrado (incluye "Otros", que al ser una suma
  // podría superar al mayor país individual): así ninguna barra excede el 100%.
  const max = rows.reduce((m, r) => Math.max(m, r.value), 0);

  return (
    <Card className="flex flex-col">
      <div className="mb-3 flex items-center gap-2">
        <span className="text-brand-600 dark:text-brand-300" aria-hidden>
          {icon}
        </span>
        <h3 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
          {title}
        </h3>
      </div>

      {rows.length === 0 ? (
        <p className="py-6 text-center text-sm text-zinc-400">{emptyLabel}</p>
      ) : (
        <ul className="space-y-2.5">
          {rows.map((row) => {
            const pct = max > 0 ? Math.round((row.value / max) * 100) : 0;
            return (
              <li key={row.key}>
                <div className="mb-1 flex items-center justify-between gap-2 text-sm">
                  <span
                    className="flex min-w-0 items-center gap-2 text-zinc-700 dark:text-zinc-200"
                    title={row.label}
                  >
                    <span
                      className="size-2.5 shrink-0 rounded-full"
                      style={{ backgroundColor: row.color }}
                      aria-hidden
                    />
                    <span className="truncate">{row.label}</span>
                  </span>
                  <span className="shrink-0 tabular-nums text-zinc-500 dark:text-zinc-400">
                    {formatNumber(row.value)}
                  </span>
                </div>
                <div className="h-1.5 w-full overflow-hidden rounded-full bg-zinc-100 dark:bg-zinc-800">
                  <div
                    className="h-full rounded-full"
                    style={{ width: `${pct}%`, backgroundColor: row.color }}
                  />
                </div>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}
