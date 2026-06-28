import { ArrowRight } from "lucide-react";
import { CtaLink } from "./CtaLink";
import { GridBackground } from "./GridBackground";
import { Reveal } from "./Reveal";


export function FinalCta() {
  return (
    <section className="relative overflow-hidden bg-white py-20 sm:py-28 dark:bg-zinc-950">
      <GridBackground />
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-1/2 -z-0 h-64 w-[36rem] -translate-x-1/2 -translate-y-1/2 rounded-full bg-brand-600/15 blur-3xl dark:bg-brand-600/20"
      />

      <Reveal className="relative mx-auto max-w-2xl px-4 text-center sm:px-6">
        <h2 className="text-balance text-3xl font-bold tracking-tight text-zinc-900 sm:text-4xl dark:text-white">
          Tu próximo enlace corto te está esperando.
        </h2>
        <p className="mx-auto mt-4 max-w-md text-lg text-zinc-600 dark:text-zinc-400">
          Crea tu cuenta gratis y comparte tu primer enlace en segundos.
        </p>
        <div className="mt-8 flex justify-center">
          <CtaLink href="/register" variant="primary" className="h-12 px-7 text-base">
            Crear cuenta gratis
            <ArrowRight className="size-4" aria-hidden />
          </CtaLink>
        </div>
      </Reveal>
    </section>
  );
}
