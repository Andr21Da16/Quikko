"use client"; // Los Error Boundaries deben ser Client Components (requisito de Next.js).

// global-error.tsx (convención de Next.js App Router): captura errores del PROPIO root
// layout (app/layout.tsx) — el único caso que app/error.tsx no puede cubrir, ya que
// error.tsx no envuelve al layout por encima suyo. 
//
// Cuando se activa, este archivo REEMPLAZA al root layout: no hay ThemeProvider, ni
// fuente, ni globals.css heredados, así que debe declarar su propio <html>/<body> e
// importar lo que necesite. Es deliberadamente autocontenido y minimal: si el root layout
// falló, conviene depender de lo menos posible (no reutilizamos ErrorScreen/GridBackground
// para no arrastrar dependencias que podrían estar implicadas en el fallo).
//
// Sin ThemeProvider no hay forma de leer el tema, así que se fija el look dark-first de la
// marca (coherente con la identidad de Quikko) — es un fallback extremo y poco frecuente.
import "./globals.css";
import { spaceGrotesk } from "./fonts";

export default function GlobalError({
  reset,
  unstable_retry,
}: {
  error: Error & { digest?: string };
  reset: () => void;
  unstable_retry?: () => void;
}) {
  const retry = unstable_retry ?? reset;

  return (
    // global-error debe incluir sus propias etiquetas <html> y <body>.
    <html lang="es" className={`${spaceGrotesk.variable} dark`}>
      <body className="flex min-h-dvh flex-col items-center justify-center bg-zinc-950 px-4 py-12 text-center text-zinc-100 antialiased">
        <span className="text-2xl font-bold tracking-tight text-zinc-50">
          Quikko<span className="text-accent-400">.</span>
        </span>
        <h1 className="mt-8 text-2xl font-bold tracking-tight text-zinc-50">
          Algo salió mal
        </h1>
        <p className="mt-2 max-w-md text-sm text-zinc-400">
          Ocurrió un error inesperado. Intenta de nuevo o recarga la página.
        </p>
        <button
          type="button"
          onClick={() => retry()}
          className="mt-8 inline-flex h-11 items-center justify-center rounded-lg bg-brand-600 px-5 text-sm font-medium text-white transition-colors hover:bg-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
        >
          Reintentar
        </button>
      </body>
    </html>
  );
}
