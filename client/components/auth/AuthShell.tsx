import Link from "next/link";
import type { ReactNode } from "react";
import { Card } from "@/components/ui";
import { GridBackground } from "@/components/landing/GridBackground";

// Shell visual compartido de las páginas de auth. Reutiliza el GridBackground
// de la landing como fondo "blueprint" tenue —se presta bien fuera del hero,
// es neutro y respeta dark/light— y centra una Card con el wordmark de Quikko arriba.

export function AuthShell({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <main className="relative flex min-h-dvh items-center justify-center overflow-hidden bg-white px-4 py-12 dark:bg-zinc-950">
      <GridBackground />

      <div className="relative z-10 w-full max-w-md">
        <div className="mb-8 flex justify-center">
          <Link
            href="/"
            className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50"
          >
            Quikko<span className="text-accent-400">.</span>
          </Link>
        </div>

        <Card className="p-8">
          <header className="mb-6 text-center">
            <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-50">
              {title}
            </h1>
            {subtitle && (
              <p className="mt-1.5 text-sm text-zinc-500 dark:text-zinc-400">
                {subtitle}
              </p>
            )}
          </header>

          {children}
        </Card>

        {footer && (
          <p className="mt-6 text-center text-sm text-zinc-500 dark:text-zinc-400">
            {footer}
          </p>
        )}
      </div>
    </main>
  );
}
