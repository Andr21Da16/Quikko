"use client";

// Overview del dashboard  — el "home" que ve el usuario al loguearse.
//

import { useEffect, useState } from "react";
import { authApi } from "@/lib/api/endpoints/auth";
import { useAuthStore } from "@/store/auth";
import { useUrlsStore } from "@/store/urls";
import { useUIStore } from "@/store/ui";
import { useRealtimeStore } from "@/store/realtime";
import type { AccountSummary } from "@/types";
import { AccountSummaryCard } from "@/components/dashboard/overview/AccountSummaryCard";
import { QuickCreateAction } from "@/components/dashboard/overview/QuickCreateAction";
import { RecentUrlsList } from "@/components/dashboard/overview/RecentUrlsList";
import { WelcomeChecklist } from "@/components/dashboard/onboarding/WelcomeChecklist";

// Cuántas URLs recientes se muestran en el vistazo (no la tabla completa).
const RECENT_LIMIT = 5;

export default function DashboardOverviewPage() {
  const userEmail = useAuthStore((s) => s.user?.email);

  // Resumen de cuenta: vía authApi (capa de datos sancionada), tal como pide la spec.

  const [summary, setSummary] = useState<AccountSummary | null>(null);
  const [summaryLoading, setSummaryLoading] = useState(true);

  // URLs recientes: vía el store de URLs. Pedimos la primera página y
  // mostramos solo las primeras RECENT_LIMIT.
  const urls = useUrlsStore((s) => s.urls);
  const urlsLoading = useUrlsStore((s) => s.isLoading);
  const fetchUrls = useUrlsStore((s) => s.fetchUrls);
  const pagination = useUrlsStore((s) => s.pagination);

  // Onboarding : "usuario nuevo" = cero URLs (derivado de datos reales, no un
  // flag aparte). Solo decidimos una vez cargada la paginación, para no parpadear; y se
  // respeta la preferencia de "ocultar" persistida en store/ui.ts. Si el usuario crea su
  // primera URL, `urls.length`/`total` dejan de ser 0 y el banner desaparece sin recargar.
  const onboardingDismissed = useUIStore((s) => s.onboardingDismissed);
  const dismissOnboarding = useUIStore((s) => s.dismissOnboarding);
  const isNewUser =
    pagination !== null && pagination.total === 0 && urls.length === 0;
  const showOnboarding = isNewUser && !onboardingDismissed;

  // Clics en vivo (WebSocket, canal user:{id}). Contador global desde que se abrió la
  // página + desglose por shortCode para el micro-indicador de cada fila.
  const [liveTotal, setLiveTotal] = useState(0);
  const [liveByCode, setLiveByCode] = useState<Record<string, number>>({});

  // Carga inicial de datos reales. (summaryLoading ya arranca en true, no hace falta
  // re-setearlo síncronamente aquí.)
  useEffect(() => {
    let active = true;
    authApi
      .getAccountSummary()
      .then((data) => {
        if (active) setSummary(data);
      })
      .catch(() => {
        // El error 401/refresh ya lo maneja el cliente HTTP; aquí solo dejamos el
        // estado "no se pudo cargar" que la tarjeta muestra con su fallback.
        if (active) setSummary(null);
      })
      .finally(() => {
        if (active) setSummaryLoading(false);
      });

    void fetchUrls(1);

    return () => {
      active = false;
    };
  }, [fetchUrls]);

  // Realtime: aseguramos la conexión (idempotente) y reflejamos cada clic nuevo SIN
  // refetch — solo incrementamos contadores. El store appendea un click_event a la vez,
  // así que detectamos el recién añadido comparando el último elemento entre estados.
  useEffect(() => {
    useRealtimeStore.getState().connect();

    const unsub = useRealtimeStore.subscribe((state, prev) => {
      const ev = state.lastEvents[state.lastEvents.length - 1];
      const prevEv = prev.lastEvents[prev.lastEvents.length - 1];
      if (!ev || ev === prevEv) return;

      setLiveTotal((n) => n + 1);
      // Solo desglosamos por código si la URL afectada está visible en la lista (leemos
      // el estado más reciente del store de URLs, no una copia capturada al montar).
      const visible = useUrlsStore
        .getState()
        .urls.slice(0, RECENT_LIMIT)
        .some((u) => u.shortCode === ev.shortCode);
      if (visible) {
        setLiveByCode((m) => ({
          ...m,
          [ev.shortCode]: (m[ev.shortCode] ?? 0) + 1,
        }));
      }
    });

    return unsub;
  }, []);

  const recentUrls = urls.slice(0, RECENT_LIMIT);
  const greeting = userEmail ? userEmail.split("@")[0] : null;

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          {greeting ? `Hola, ${greeting}` : "Resumen"}
        </h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Un vistazo a tus enlaces y su actividad reciente.
        </p>
      </header>

      {showOnboarding && <WelcomeChecklist onDismiss={dismissOnboarding} />}

      <AccountSummaryCard
        summary={summary}
        liveClicks={liveTotal}
        isLoading={summaryLoading}
      />

      <QuickCreateAction />

      <RecentUrlsList
        urls={recentUrls}
        liveByCode={liveByCode}
        isLoading={urlsLoading}
      />
    </div>
  );
}
