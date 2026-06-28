import type { ReactNode } from "react";
import Link from "next/link";
import { GridBackground } from "@/components/landing/GridBackground";

// Pantalla de error a pantalla completa, compartida por la 404 (app/not-found.tsx) y la
// página de error de runtime (app/error.tsx). Presentacional y SIN hooks: por eso puede
// renderizarse tanto desde un Server Component (not-found) como dentro del boundary de
// cliente (error). Mismo lenguaje visual que la landing y las páginas de auth: el
// GridBackground "blueprint" tenue + el wordmark Quikko, tokens de marca y dark mode.
export function ErrorScreen({
  code,
  title,
  message,
  actions,
}: {
  /** Elemento visual destacado (ej. "404" con gradiente, o un ícono en badge). */
  code: ReactNode;
  title: string;
  message: string;
  actions: ReactNode;
}) {
  return (
    <main className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-white px-4 py-12 text-center dark:bg-zinc-950">
      <GridBackground />

      <div className="relative z-10 flex w-full max-w-md flex-col items-center">
        <Link
          href="/"
          className="mb-8 text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50"
        >
          Quikko<span className="text-accent-400">.</span>
        </Link>

        {code}

        <h1 className="mt-4 text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          {title}
        </h1>
        <p className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">{message}</p>

        <div className="mt-8 flex flex-col gap-3 sm:flex-row">{actions}</div>
      </div>
    </main>
  );
}
