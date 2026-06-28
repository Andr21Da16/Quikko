import { Check } from "lucide-react";
import { cn } from "@/lib/utils/cn";
import { CtaLink } from "./CtaLink";
import { Reveal } from "./Reveal";

// Pricing. Free vs Pro SIN precios (no hay pasarela de pago todavía):
// se usan frases de posicionamiento en vez de un número. La tarjeta Pro se distingue
// (borde de acento + badge "Más usado") sin verse completamente distinta a la Free.
type Tier = {
  name: string;
  tagline: string;
  features: string[];
  cta: string;
  featured?: boolean;
};

const TIERS: Tier[] = [
  {
    name: "Free",
    tagline: "Para empezar",
    features: [
      "Hasta 5 URLs activas",
      "Analytics de las últimas 24 h",
      "Código QR incluido en cada enlace",
      "Dashboard en tiempo real",
    ],
    cta: "Crear cuenta gratis",
  },
  {
    name: "Pro",
    tagline: "Para crecer sin límites",
    features: [
      "URLs activas ilimitadas",
      "Histórico de 7 y 30 días",
      "Export de métricas a CSV",
      "Todo lo del plan Free",
    ],
    cta: "Empezar con Pro",
    featured: true,
  },
];

export function Pricing() {
  return (
    <section
      id="pricing"
      className="border-b border-zinc-200 bg-white py-20 sm:py-28 dark:border-zinc-900 dark:bg-zinc-950"
    >
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <Reveal className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-zinc-900 sm:text-4xl dark:text-white">
            Un plan para cada etapa
          </h2>
          <p className="mt-4 text-lg text-zinc-600 dark:text-zinc-400">
            Empieza gratis y mejora cuando tus enlaces despeguen.
          </p>
        </Reveal>

        <div className="mx-auto mt-14 grid max-w-3xl grid-cols-1 gap-6 md:grid-cols-2">
          {TIERS.map((tier, i) => (
            <Reveal key={tier.name} delay={i * 0.08}>
              <div
                className={cn(
                  "relative flex h-full flex-col rounded-2xl border bg-white p-6 dark:bg-zinc-950",
                  tier.featured
                    ? "border-accent-400 shadow-lg shadow-accent-400/10 dark:border-accent-500/70"
                    : "border-zinc-200 dark:border-zinc-800",
                )}
              >
                {tier.featured && (
                  <span className="absolute -top-3 right-6 rounded-full bg-accent-400 px-3 py-0.5 text-xs font-semibold text-zinc-900">
                    Más usado
                  </span>
                )}

                <h3 className="text-lg font-bold text-zinc-900 dark:text-white">
                  {tier.name}
                </h3>
                <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                  {tier.tagline}
                </p>

                <ul className="mt-6 flex flex-1 flex-col gap-3">
                  {tier.features.map((feature) => (
                    <li
                      key={feature}
                      className="flex items-start gap-2.5 text-sm text-zinc-700 dark:text-zinc-300"
                    >
                      <Check
                        className={cn(
                          "mt-0.5 size-4 shrink-0",
                          tier.featured
                            ? "text-accent-600 dark:text-accent-400"
                            : "text-brand-600 dark:text-brand-400",
                        )}
                        aria-hidden
                      />
                      {feature}
                    </li>
                  ))}
                </ul>

                <CtaLink
                  href="/register"
                  variant={tier.featured ? "primary" : "secondary"}
                  className="mt-8 w-full"
                >
                  {tier.cta}
                </CtaLink>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
