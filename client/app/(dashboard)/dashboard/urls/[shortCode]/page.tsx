"use client";

// Detalle de una URL  — vista profunda de un solo enlace con métricas en vivo.
//

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft } from "lucide-react";
import { Spinner } from "@/components/ui";
import { urlsApi } from "@/lib/api/endpoints/urls";
import { analyticsApi } from "@/lib/api/endpoints/analytics";
import { authApi } from "@/lib/api/endpoints/auth";
import { useUrlsStore } from "@/store/urls";
import { useRealtimeStore } from "@/store/realtime";
import { useNotificationsStore } from "@/store/notifications";
import { ApiError } from "@/lib/api/client";
import type { ClickStats, Plan, ShortURL, TimeRange } from "@/types";
import { UrlDetailHeader } from "@/components/dashboard/url-detail/UrlDetailHeader";
import { RangeSelector } from "@/components/dashboard/charts/RangeSelector";
import { StatsPanels } from "@/components/dashboard/charts/StatsPanels";

const EMPTY_STATS: ClickStats = {
  totalClicks: 0,
  clicksByCountry: {},
  clicksByDevice: {},
  clicksByBrowser: {},
  clicksOverTime: [],
};

function BackLink() {
  return (
    <Link
      href="/dashboard/urls"
      className="inline-flex items-center gap-1.5 text-sm font-medium text-zinc-500 transition-colors hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-100"
    >
      <ArrowLeft className="size-4" aria-hidden />
      Mis URLs
    </Link>
  );
}

export default function UrlDetailPage() {
  const params = useParams<{ shortCode: string }>();
  const shortCode = Array.isArray(params.shortCode)
    ? params.shortCode[0]
    : params.shortCode;

  const toggleActive = useUrlsStore((s) => s.toggleActive);
  const notify = useNotificationsStore((s) => s.notify);
  const isConnected = useRealtimeStore((s) => s.isConnected);

  const [url, setUrl] = useState<ShortURL | null>(null);
  const [urlError, setUrlError] = useState<string | null>(null);
  const [plan, setPlan] = useState<Plan | null>(null);
  const [range, setRange] = useState<TimeRange>("24h");
  // Guardamos junto al stat el rango al que pertenece: así "cargando" se deriva de que el
  // rango del resultado no coincida con el seleccionado, sin un setState síncrono en effect.
  const [statsResult, setStatsResult] = useState<{
    range: TimeRange;
    data: ClickStats;
  } | null>(null);
  const [livePulse, setLivePulse] = useState(false);

  const urlReady = url !== null;
  const statsLoading = !statsResult || statsResult.range !== range;
  const stats = statsResult?.data ?? EMPTY_STATS;

  // --- Carga de metadatos de la URL + plan ---
  const loadUrl = useCallback(
    () =>
      urlsApi
        .getURLByCode(shortCode)
        .then((u) => {
          setUrl(u);
          setUrlError(null);
        })
        .catch((err) =>
          setUrlError(err instanceof ApiError ? err.code : "GENERIC"),
        ),
    [shortCode],
  );

  const refreshPlan = useCallback(() => {
    authApi
      .getAccountSummary()
      .then((s) => setPlan(s.plan))
      .catch(() => setPlan(null));
  }, []);

  useEffect(() => {
    void loadUrl();
    refreshPlan();
  }, [loadUrl, refreshPlan]);

  // --- Carga de stats por rango (solo una vez que la URL existe) ---
  const statsToken = useRef(0);
  const loadStats = useCallback(
    (r: TimeRange) => {
      const token = ++statsToken.current;
      return analyticsApi
        .getStats({ shortCode, range: r })
        .then((res) => {
          if (token === statsToken.current)
            setStatsResult({ range: r, data: res.stats });
        })
        .catch(() => {
          // El selector ya impide rangos no permitidos; ante cualquier fallo mostramos
          // vacío en vez de romper.
          if (token === statsToken.current)
            setStatsResult({ range: r, data: EMPTY_STATS });
        });
    },
    [shortCode],
  );

  useEffect(() => {
    if (!urlReady) return;
    void loadStats(range);
  }, [urlReady, range, loadStats]);

  // --- Tiempo real: suscribirse al canal url:{shortCode} y limpiar al salir ---
  useEffect(() => {
    useRealtimeStore.getState().connect(); // idempotente
  }, []);

  // (Re)suscribir cuando la conexión esté abierta (el mensaje de subscribe se descarta si
  // el socket aún no está OPEN). Al desmontar o reconectar/cambiar de URL, desuscribir.
  useEffect(() => {
    if (!isConnected) return;
    const rt = useRealtimeStore.getState();
    rt.subscribeToUrl(shortCode);
    return () => rt.unsubscribeFromUrl(shortCode);
  }, [isConnected, shortCode]);

  // Indicador "nuevo clic" momentáneo.
  const pulseTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  useEffect(() => {
    return () => {
      if (pulseTimer.current) clearTimeout(pulseTimer.current);
    };
  }, []);

  // Procesar cada click_event nuevo de esta URL SIN refetch: incrementar contadores +
  // desgloses + el último bucket del gráfico, y pulsar el indicador.
  useEffect(() => {
    const unsub = useRealtimeStore.subscribe((state, prev) => {
      const ev = state.lastEvents[state.lastEvents.length - 1];
      const prevEv = prev.lastEvents[prev.lastEvents.length - 1];
      if (!ev || ev === prevEv || ev.shortCode !== shortCode) return;

      setUrl((u) => (u ? { ...u, totalClicks: u.totalClicks + 1 } : u));
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
  }, [shortCode]);

  // --- Toggle activar/desactivar (vía store, con espejo local optimista) ---
  const handleToggle = async () => {
    if (!url) return;
    const next = !url.isActive;
    setUrl({ ...url, isActive: next });
    try {
      await toggleActive(url.id, next);
    } catch {
      setUrl({ ...url, isActive: !next });
      notify("error", "No se pudo cambiar el estado de la URL.");
    }
  };

  // --- Estados de página ---
  if (urlError) {
    return (
      <div className="mx-auto max-w-6xl space-y-6">
        <BackLink />
        <div className="flex flex-col items-center gap-3 rounded-xl border border-zinc-200 bg-white px-6 py-16 text-center dark:border-zinc-800 dark:bg-zinc-900">
          <p className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
            URL no encontrada
          </p>
          <p className="max-w-sm text-sm text-zinc-500 dark:text-zinc-400">
            Esta URL no existe o no pertenece a tu cuenta.
          </p>
          <Link
            href="/dashboard/urls"
            className="mt-1 inline-flex h-10 items-center gap-2 rounded-lg bg-brand-600 px-4 text-sm font-medium text-white transition-colors hover:bg-brand-700"
          >
            Volver a Mis URLs
          </Link>
        </div>
      </div>
    );
  }

  if (!url) {
    return (
      <div className="mx-auto max-w-6xl space-y-6">
        <BackLink />
        <div className="flex justify-center py-24">
          <Spinner size={28} className="text-brand-600 dark:text-brand-400" />
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <BackLink />

      <UrlDetailHeader url={url} onToggle={handleToggle} livePulse={livePulse} />

      <div className="flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
          Métricas
        </h2>
        <RangeSelector value={range} onChange={setRange} plan={plan} />
      </div>

      {statsLoading ? (
        <div className="flex justify-center rounded-xl border border-zinc-200 bg-white py-24 dark:border-zinc-800 dark:bg-zinc-900">
          <Spinner size={24} className="text-brand-600 dark:text-brand-400" />
        </div>
      ) : (
        <StatsPanels stats={stats} range={range} />
      )}
    </div>
  );
}
