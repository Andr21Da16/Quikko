import { ClipboardPaste, Share2, Activity } from "lucide-react";
import { Reveal } from "./Reveal";
import type { LucideIcon } from "lucide-react";

// Cómo funciona: 3 pasos numerados, simple y escaneable.
type Step = { n: string; icon: LucideIcon; title: string; desc: string };

const STEPS: Step[] = [
  {
    n: "01",
    icon: ClipboardPaste,
    title: "Pega tu URL larga",
    desc: "Opcionalmente elige un alias propio. Quikko genera el enlace corto al instante.",
  },
  {
    n: "02",
    icon: Share2,
    title: "Comparte el corto y su QR",
    desc: "Cada enlace trae su código QR listo para web, impresos o redes.",
  },
  {
    n: "03",
    icon: Activity,
    title: "Ve los clics llegar en vivo",
    desc: "Tu dashboard se actualiza en tiempo real con cada clic: país, dispositivo y navegador.",
  },
];

export function HowItWorks() {
  return (
    <section
      id="how-it-works"
      className="border-b border-zinc-200 bg-zinc-50 py-20 sm:py-28 dark:border-zinc-900 dark:bg-zinc-900/30"
    >
      <div className="mx-auto max-w-6xl px-4 sm:px-6">
        <Reveal className="mx-auto max-w-2xl text-center">
          <h2 className="text-3xl font-bold tracking-tight text-zinc-900 sm:text-4xl dark:text-white">
            Cómo funciona
          </h2>
          <p className="mt-4 text-lg text-zinc-600 dark:text-zinc-400">
            De una URL larga a métricas en vivo en tres pasos.
          </p>
        </Reveal>

        <div className="mt-14 grid grid-cols-1 gap-6 md:grid-cols-3">
          {STEPS.map((step, i) => {
            const Icon = step.icon;
            return (
              <Reveal key={step.n} delay={i * 0.08}>
                <div className="relative h-full rounded-2xl border border-zinc-200 bg-white p-6 dark:border-zinc-800 dark:bg-zinc-950">
                  <span className="text-sm font-bold text-accent-500 dark:text-accent-400">
                    {step.n}
                  </span>
                  <span className="mt-3 flex size-10 items-center justify-center rounded-lg bg-brand-50 text-brand-600 dark:bg-brand-900/40 dark:text-brand-300">
                    <Icon className="size-5" aria-hidden />
                  </span>
                  <h3 className="mt-4 font-semibold text-zinc-900 dark:text-white">
                    {step.title}
                  </h3>
                  <p className="mt-1.5 text-sm leading-relaxed text-zinc-600 dark:text-zinc-400">
                    {step.desc}
                  </p>
                </div>
              </Reveal>
            );
          })}
        </div>
      </div>
    </section>
  );
}
