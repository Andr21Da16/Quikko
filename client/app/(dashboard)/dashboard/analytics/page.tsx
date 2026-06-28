"use client";

// Analytics — métricas AGREGADAS de todas las URLs del usuario. 
// Reutiliza los gráficos compartidos (StatsPanels:
// serie temporal D3 + desgloses) y el RangeSelector con gating de plan; aquí además vive
// la exportación a CSV.
//
// TIEMPO REAL: el backend solo expone los canales user:{id} y url:{code} (no un canal
// "agregado" propio), PERO user:{id} —al que el servidor une automáticamente al conectar—
// entrega TODOS los clics del usuario, así que sirve como fuente agregada en vivo. Por eso
// esta página SÍ actualiza en vivo (incrementa contadores sin refetch), igual que Overview.

import { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Radio } from "lucide-react";
import { Spinner } from "@/components/ui";
import { analyticsApi } from "@/lib/api/endpoints/analytics";
import { authApi } from "@/lib/api/endpoints/auth";
import { useUrlsStore } from "@/store/urls";
import { useRealtimeStore } from "@/store/realtime";
import type { ClickStats, Plan, TimeRange } from "@/types";
import { RangeSelector } from "@/components/dashboard/charts/RangeSelector";
import { StatsPanels } from "@/components/dashboard/charts/StatsPanels";
import { ExportCsvButton } from "@/components/dashboard/analytics/ExportCsvButton";
import { TopUrlsCard } from "@/components/dashboard/analytics/TopUrlsCard";

const EMPTY_STATS: ClickStats = {
  totalClicks: 0,
  clicksByCountry: {},
  clicksByDevice: {},
  clicksByBrowser: {},
  clicksOverTime: [],
};

export default function AnalyticsPage() {
  const urls = useUrlsStore((s) => s.urls);
  const fetchUrls = useUrlsStore((s) => s.fetchUrls);

  const [plan, setPlan] = useState<Plan | null>(null);
  const [range, setRange] = useState<TimeRange>("24h");
  const [statsResult, setStatsResult] = useState<{
    range: TimeRange;
    data: ClickStats;
  } | null>(null);
  const [livePulse, setLivePulse] = useState(false);

  const statsLoading = !statsResult || statsResult.range !== range;
  const stats = statsResult?.data ?? EMPTY_STATS;

  // --- Plan (para el gating de rango) + URLs (para el ranking) ---
  const refreshPlan = useCallback(() => {
    authApi
      .getAccountSummary()
      .then((s) => setPlan(s.plan))
      .catch(() => setPlan(null));
  }, []);

  useEffect(() => {
    refreshPlan();
    void fetchUrls(1);
  }, [refreshPlan, fetchUrls]);

  // --- Stats agregadas por rango (getStats SIN shortCode) ---
  const statsToken = useRef(0);
  const loadStats = useCallback((r: TimeRange) => {
    const token = ++statsToken.current;
    return analyticsApi
      .getStats({ range: r })
      .then((res) => {
        if (token === statsToken.current)
          setStatsResult({ range: r, data: res.stats });
      })
      .catch(() => {
        if (token === statsToken.current)
          setStatsResult({ range: r, data: EMPTY_STATS });
      });
  }, []);

  useEffect(() => {
    void loadStats(range);
  }, [range, loadStats]);

  // --- Tiempo real: canal user:{id} (auto-unido al conectar) → agregado en vivo ---
  useEffect(() => {
    useRealtimeStore.getState().connect(); // idempotente
  }, []);

  const pulseTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    return () => {
      if (pulseTimer.current) clearTimeout(pulseTimer.current);
    };
  }, []);

  useEffect(() => {
    const unsub = useRealtimeStore.subscribe((state, prev) => {
      const ev = state.lastEvents[state.lastEvents.length - 1];
      const prevEv = prev.lastEvents[prev.lastEvents.length - 1];
      if (!ev || ev === prevEv) return; // cualquier clic del usuario cuenta (agregado)

      setStatsResult((sr) => {
        if (!sr) return sr;
        const d = sr.data;
        const bump = (m: Record<string, number>, k: string) =>
          k ? { ...m, [k]: (m[k] ?? 0) + 1 } : m;
        const cot = d.clicksOverTime.length
          ? d.clicksOverTime.map((b, i) =>
              i === d.clicksOverTime.length - 1
                ? { ...b, count: b.count + 1 }
                : b,
            )
          : d.clicksOverTime;
        return {
          range: sr.range,
          data: {
            ...d,
            totalClicks: d.totalClicks + 1,
            clicksByCountry: bump(d.clicksByCountry, ev.country),
            clicksByDevice: bump(d.clicksByDevice, ev.deviceType),
            clicksByBrowser: bump(d.clicksByBrowser, ev.browser),
            clicksOverTime: cot,
          },
        };
      });

      setLivePulse(true);
      if (pulseTimer.current) clearTimeout(pulseTimer.current);
      pulseTimer.current = setTimeout(() => setLivePulse(false), 1500);
    });
    return unsub;
  }, []);

  const hasData = stats.totalClicks > 0;

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <header className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h1 className="flex items-center gap-2 text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
            Analytics
            {livePulse && (
              <span className="inline-flex items-center gap-1 rounded-full bg-accent-100 px-1.5 py-0.5 text-[10px] font-semibold text-accent-700 dark:bg-accent-900/40 dark:text-accent-300">
                <Radio className="size-2.5 animate-pulse" aria-hidden />
                en vivo
              </span>
            )}
          </h1>
          <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
            Métricas combinadas de todas tus URLs.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <RangeSelector value={range} onChange={setRange} plan={plan} />
          <ExportCsvButton range={range} />
        </div>
      </header>

      {statsLoading ? (
        <div className="flex justify-center rounded-xl border border-zinc-200 bg-white py-24 dark:border-zinc-800 dark:bg-zinc-900">
          <Spinner size={24} className="text-brand-600 dark:text-brand-400" />
        </div>
      ) : hasData ? (
        <>
          <StatsPanels stats={stats} range={range} />
          <TopUrlsCard urls={urls} />
        </>
      ) : (
        <div className="flex flex-col items-center gap-3 rounded-xl border border-zinc-200 bg-white px-6 py-16 text-center dark:border-zinc-800 dark:bg-zinc-900">
          <p className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
            Aún no hay clics en este rango
          </p>
          <p className="max-w-sm text-sm text-zinc-500 dark:text-zinc-400">
            Cuando tus enlaces empiecen a recibir visitas, verás aquí las
            métricas agregadas de todas tus URLs.
          </p>
          <Link
            href="/dashboard/urls"
            className="mt-1 inline-flex h-10 items-center gap-2 rounded-lg bg-brand-600 px-4 text-sm font-medium text-white transition-colors hover:bg-brand-700"
          >
            Ir a Mis URLs
          </Link>
        </div>
      )}
    </div>
  );
}
