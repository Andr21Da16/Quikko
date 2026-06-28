import Link from "next/link";
import { ArrowRight, Link2, QrCode, Radio, Sparkles, X } from "lucide-react";
import { Card } from "@/components/ui";

// Banner de bienvenida / checklist de onboarding (Agent 28).
//
// DECISIÓN DE DISEÑO: "usuario nuevo" = cero URLs creadas, condición derivada de datos
// reales (la paginación del store de URLs), no un flag separado que haya que sincronizar.
// En consecuencia el banner desaparece de forma natural en cuanto el usuario crea su
// primera URL (al volver a Overview se refetchea y `total` deja de ser 0): no se "marca
// completado" porque el banner ya no aplica. Es una versión más rica del estado vacío de
// Overview, presentada arriba como guía de primeros pasos. El usuario también puede
// ocultarlo manualmente (preferencia persistida en store/ui.ts) aunque siga sin URLs.
//
// Se optó por banner/checklist en vez de un tour interactivo con tooltips (requeriría una
// librería tipo react-joyride)

type Step = {
  icon: typeof Link2;
  title: string;
  description: string;
  href?: string;
  cta?: string;
};

const STEPS: Step[] = [
  {
    icon: Link2,
    title: "Crea tu primer enlace corto",
    description:
      "Pega una URL larga y obtén un enlace Quikko listo para compartir.",
    href: "/dashboard/urls",
    cta: "Crear URL",
  },
  {
    icon: QrCode,
    title: "Compártelo y muestra el QR",
    description: "Cada enlace genera su código QR automáticamente al crearlo.",
  },
  {
    icon: Radio,
    title: "Mira los clics en tiempo real",
    description:
      "El dashboard se actualiza solo: verás cada clic llegar al instante.",
  },
  {
    icon: Sparkles,
    title: "Conoce los planes Free y Pro",
    description: "Amplía tus límites y desbloquea rangos de fecha más largos.",
    href: "/dashboard/account",
    cta: "Ver planes",
  },
];

export function WelcomeChecklist({ onDismiss }: { onDismiss: () => void }) {
  return (
    <Card className="relative overflow-hidden border-brand-200 bg-gradient-to-br from-brand-50 to-white dark:border-brand-900/50 dark:from-brand-950/40 dark:to-zinc-900">
      <button
        type="button"
        onClick={onDismiss}
        aria-label="Ocultar guía de bienvenida"
        className="absolute right-3 top-3 inline-flex size-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 dark:hover:bg-zinc-800 dark:hover:text-zinc-300"
      >
        <X className="size-4" aria-hidden />
      </button>

      <div className="max-w-xl pr-8">
        <span className="inline-flex items-center gap-1.5 rounded-full bg-brand-100 px-2.5 py-1 text-xs font-semibold text-brand-700 dark:bg-brand-900/50 dark:text-brand-300">
          <Sparkles className="size-3.5" aria-hidden />
          Primeros pasos
        </span>
        <h2 className="mt-3 text-lg font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          ¡Te damos la bienvenida a Quikko!
        </h2>
        <p className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
          Crea tu primer enlace y empieza a medir clics en segundos.
        </p>
      </div>

      <ol className="mt-5 grid gap-3 sm:grid-cols-2">
        {STEPS.map((step, i) => {
          const Icon = step.icon;
          return (
            <li
              key={step.title}
              className="flex gap-3 rounded-xl border border-zinc-200 bg-white/70 p-4 dark:border-zinc-800 dark:bg-zinc-900/60"
            >
              <span className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-brand-100 text-brand-700 dark:bg-brand-900/50 dark:text-brand-300">
                <Icon className="size-5" aria-hidden />
              </span>
              <div className="min-w-0">
                <p className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
                  <span className="text-brand-600 dark:text-brand-400">
                    {i + 1}.
                  </span>{" "}
                  {step.title}
                </p>
                <p className="mt-0.5 text-xs text-zinc-500 dark:text-zinc-400">
                  {step.description}
                </p>
                {step.href && step.cta && (
                  <Link
                    href={step.href}
                    className="mt-2 inline-flex items-center gap-1 text-xs font-semibold text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
                  >
                    {step.cta}
                    <ArrowRight className="size-3" aria-hidden />
                  </Link>
                )}
              </div>
            </li>
          );
        })}
      </ol>
    </Card>
  );
}
