"use client";

// Cuenta — gestión de la cuenta: resumen, cambio de plan (sin pago), cambio de
// contraseña y eliminación de cuenta (con confirmación fuerte). Reutiliza AccountSummaryCard
// de Overview para el resumen de uso, en vez de duplicarlo.

import { useCallback, useEffect, useState } from "react";
import { authApi } from "@/lib/api/endpoints/auth";
import { useRealtimeStore } from "@/store/realtime";
import { useUrlsStore } from "@/store/urls";
import type { AccountSummary } from "@/types";
import { AccountSummaryCard } from "@/components/dashboard/overview/AccountSummaryCard";
import { AccountInfoCard } from "@/components/dashboard/account/AccountInfoCard";
import { PlanSection } from "@/components/dashboard/account/PlanSection";
import { ChangePasswordForm } from "@/components/dashboard/account/ChangePasswordForm";
import { DangerZone } from "@/components/dashboard/account/DangerZone";

export default function AccountPage() {
  const [summary, setSummary] = useState<AccountSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [liveClicks, setLiveClicks] = useState(0);

  // Refresca el resumen tras un cambio de plan (cambia el cupo/limit) o al montar.
  const refreshSummary = useCallback(() => {
    authApi
      .getAccountSummary()
      .then((s) => {
        setSummary(s);
        setLoading(false);
      })
      .catch(() => {
        setSummary(null);
        setLoading(false);
      });
  }, []);

  useEffect(() => {
    refreshSummary();
  }, [refreshSummary]);

  // El cambio de plan también afecta el cupo en "Mis URLs": refrescamos esa lista para
  // que el banner de cupo se recalcule si el usuario vuelve allí.
  const fetchUrls = useUrlsStore((s) => s.fetchUrls);
  const handlePlanChanged = useCallback(() => {
    refreshSummary();
    void fetchUrls(1);
  }, [refreshSummary, fetchUrls]);

  // Resumen de uso reutilizado de Overview incluye una tarjeta de "clics en vivo": la
  // alimentamos con el canal user:{id} (auto-unido al conectar) para que sea real aquí
  // también, no un 0 fijo.
  useEffect(() => {
    useRealtimeStore.getState().connect();
    const unsub = useRealtimeStore.subscribe((state, prev) => {
      const ev = state.lastEvents[state.lastEvents.length - 1];
      const prevEv = prev.lastEvents[prev.lastEvents.length - 1];
      if (!ev || ev === prevEv) return;
      setLiveClicks((n) => n + 1);
    });
    return unsub;
  }, []);

  const currentPlan = summary?.plan ?? null;

  return (
    <div className="mx-auto max-w-4xl space-y-8">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          Cuenta
        </h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Gestiona tu plan, tu contraseña y tu cuenta.
        </p>
      </header>

      <AccountInfoCard summary={summary} isLoading={loading} />

      <AccountSummaryCard
        summary={summary}
        liveClicks={liveClicks}
        isLoading={loading}
      />

      <PlanSection currentPlan={currentPlan} onChanged={handlePlanChanged} />

      <ChangePasswordForm />

      <DangerZone />
    </div>
  );
}
