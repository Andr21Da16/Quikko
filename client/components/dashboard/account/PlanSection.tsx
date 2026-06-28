"use client";

import { useState } from "react";
import { Check } from "lucide-react";
import { Badge, Button, Card } from "@/components/ui";
import { cn } from "@/lib/utils/cn";
import { useAuthStore } from "@/store/auth";
import { useNotificationsStore } from "@/store/notifications";
import { ApiError } from "@/lib/api/client";
import type { Plan } from "@/types";

// Sección de plan. Free vs Pro SIN precio (no hay pasarela; el cambio es solo
// lógico, como en el backend). Mismo copy/estructura que la sección Pricing de la landing,
//  adaptado a este contexto más compacto. Cambiar de plan pide una confirmación
// simple en línea (no un modal pesado) y actualiza el plan en el store de auth para que el
// resto de la app reaccione sin recargar.

type Tier = {
  plan: Plan;
  name: string;
  tagline: string;
  features: string[];
};

const TIERS: Tier[] = [
  {
    plan: "free",
    name: "Free",
    tagline: "Para empezar",
    features: [
      "Hasta 5 URLs activas",
      "Analytics de las últimas 24 h",
      "Código QR en cada enlace",
      "Dashboard en tiempo real",
    ],
  },
  {
    plan: "pro",
    name: "Pro",
    tagline: "Para crecer sin límites",
    features: [
      "URLs activas ilimitadas",
      "Histórico de 7 y 30 días",
      "Export de métricas a CSV",
      "Todo lo del plan Free",
    ],
  },
];

export function PlanSection({
  currentPlan,
  onChanged,
}: {
  currentPlan: Plan | null;
  onChanged?: (plan: Plan) => void;
}) {
  const updatePlan = useAuthStore((s) => s.updatePlan);
  const notify = useNotificationsStore((s) => s.notify);

  // Confirmación simple en línea: guarda el plan objetivo pendiente de confirmar.
  const [confirming, setConfirming] = useState<Plan | null>(null);
  const [saving, setSaving] = useState(false);

  const handleConfirm = async (plan: Plan) => {
    setSaving(true);
    try {
      await updatePlan(plan);
      notify("success", `Tu plan ahora es ${plan === "pro" ? "Pro" : "Free"}.`);
      setConfirming(null);
      onChanged?.(plan);
    } catch (err) {
      notify(
        "error",
        err instanceof ApiError ? err.message : "No se pudo cambiar el plan.",
      );
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
          Plan
        </h2>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Cambia de plan cuando quieras. Sin cobros: es solo una bandera lógica.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        {TIERS.map((tier) => {
          const isCurrent = currentPlan === tier.plan;
          const isConfirming = confirming === tier.plan;
          return (
            <Card
              key={tier.plan}
              className={cn(
                "flex flex-col gap-4",
                tier.plan === "pro" &&
                  "border-accent-400 dark:border-accent-500/60",
              )}
            >
              <div className="flex items-start justify-between">
                <div>
                  <div className="flex items-center gap-2">
                    <h3 className="text-base font-bold text-zinc-900 dark:text-zinc-50">
                      {tier.name}
                    </h3>
                    {isCurrent && <Badge variant="brand">Plan actual</Badge>}
                  </div>
                  <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
                    {tier.tagline}
                  </p>
                </div>
              </div>

              <ul className="flex-1 space-y-2">
                {tier.features.map((f) => (
                  <li
                    key={f}
                    className="flex items-start gap-2 text-sm text-zinc-700 dark:text-zinc-300"
                  >
                    <Check
                      className="mt-0.5 size-4 shrink-0 text-brand-600 dark:text-brand-400"
                      aria-hidden
                    />
                    {f}
                  </li>
                ))}
              </ul>

              {isCurrent ? (
                <Button variant="secondary" disabled className="w-full">
                  Tu plan actual
                </Button>
              ) : isConfirming ? (
                <div className="flex gap-2">
                  <Button
                    variant="ghost"
                    className="flex-1"
                    onClick={() => setConfirming(null)}
                    disabled={saving}
                  >
                    Cancelar
                  </Button>
                  <Button
                    className="flex-1"
                    onClick={() => handleConfirm(tier.plan)}
                    isLoading={saving}
                  >
                    Confirmar
                  </Button>
                </div>
              ) : (
                <Button
                  className="w-full"
                  onClick={() => setConfirming(tier.plan)}
                >
                  Cambiar a {tier.name}
                </Button>
              )}
            </Card>
          );
        })}
      </div>
    </section>
  );
}
