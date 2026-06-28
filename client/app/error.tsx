"use client"; // Los Error Boundaries deben ser Client Components (requisito de Next.js).

import { useEffect } from "react";
import { TriangleAlert } from "lucide-react";
import { Button } from "@/components/ui";
import { CtaLink } from "@/components/landing/CtaLink";
import { ErrorScreen } from "@/components/error/ErrorScreen";

// error.tsx (convención de Next.js App Router): Error Boundary para errores de runtime/
// renderizado en el árbol de rutas. Envuelve page/loading/not-found y los layouts
// anidados, pero NO el root layout (eso lo cubre app/global-error.tsx).
//

export default function Error({
  error,
  reset,
  unstable_retry,
}: {
  error: Error & { digest?: string };
  reset: () => void;
  unstable_retry?: () => void;
}) {
  useEffect(() => {
    // TODO: enviar a servicio de error tracking cuando exista (ej. Sentry).
    console.error(error);
  }, [error]);

  const retry = unstable_retry ?? reset;

  return (
    <ErrorScreen
      code={
        <span className="flex size-20 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-accent-500 text-white shadow-lg shadow-brand-600/30">
          <TriangleAlert className="size-10" aria-hidden />
        </span>
      }
      title="Algo salió mal"
      message="Tuvimos un problema al cargar esta parte de la app. Puedes reintentar o volver al inicio."
      actions={
        <>
          <Button onClick={() => retry()} className="h-11 px-5">
            Reintentar
          </Button>
          <CtaLink href="/dashboard" variant="secondary">
            Ir al dashboard
          </CtaLink>
        </>
      }
    />
  );
}
