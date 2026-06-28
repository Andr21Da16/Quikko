import { Zap, ArrowRight } from "lucide-react";
import { CtaLink } from "./CtaLink";
import { GridBackground } from "./GridBackground";
import { Reveal } from "./Reveal";
import { DashboardMockup } from "./DashboardMockup";

// Hero. Dark-mode-first: fondo casi negro (zinc-950) con grid blueprint y
// un glow violeta/lima detrás del mockup. Mensaje principal: velocidad/instantaneidad.
export function Hero() {
  return (
    <section className="relative overflow-hidden border-b border-zinc-200 bg-white dark:border-zinc-900 dark:bg-zinc-950">
      <GridBackground />

      <div className="relative mx-auto max-w-6xl px-4 pb-16 pt-16 sm:px-6 sm:pb-20 sm:pt-24">
        <div className="mx-auto max-w-3xl text-center">
          <Reveal>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-brand-200 bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700 dark:border-brand-800 dark:bg-brand-900/40 dark:text-brand-300">
              <Zap className="size-3.5 text-accent-500" aria-hidden />
              Redirección impulsada por Redis
            </span>
          </Reveal>

          <Reveal delay={0.05}>
            <h1 className="mt-6 text-balance text-4xl font-bold leading-[1.05] tracking-tight text-zinc-900 sm:text-5xl lg:text-6xl dark:text-white">
              Acorta enlaces. Mira los clics{" "}
              <span className="text-brand-600 dark:text-brand-400">
                llegar al instante
              </span>
              .
            </h1>
          </Reveal>

          <Reveal delay={0.1}>
            <p className="mx-auto mt-5 max-w-xl text-pretty text-lg text-zinc-600 dark:text-zinc-400">
              Quikko acorta tus URLs y te muestra cada clic en tiempo real —
              por país, dispositivo y navegador— sin esperar a un reporte.
            </p>
          </Reveal>

          <Reveal delay={0.15}>
            <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
              <CtaLink href="/register" variant="primary" className="w-full sm:w-auto">
                Crear cuenta gratis
                <ArrowRight className="size-4" aria-hidden />
              </CtaLink>
              <CtaLink
                href="#how-it-works"
                variant="secondary"
                className="w-full sm:w-auto"
              >
                Ver cómo funciona
              </CtaLink>
            </div>
          </Reveal>
        </div>

        {/* Mockup del producto con glow detrás */}
        <Reveal delay={0.2} className="relative mx-auto mt-14 max-w-4xl">
          <div
            aria-hidden
            className="absolute -inset-x-10 -top-10 -z-10 h-72 bg-brand-600/20 blur-3xl dark:bg-brand-600/25"
          />
          <div
            aria-hidden
            className="absolute -bottom-10 right-1/4 -z-10 h-40 w-40 rounded-full bg-accent-400/20 blur-3xl"
          />
          <DashboardMockup />
        </Reveal>
      </div>
    </section>
  );
}
