import type { Metadata } from "next";
import { Link2Off } from "lucide-react";
import { buildMetadata } from "@/lib/seo/metadata";
import { CtaLink } from "@/components/landing/CtaLink";
import { ErrorScreen } from "@/components/error/ErrorScreen";

// Página pública "link inactivo". El backend redirige aquí (302)
// cuando GET /:code resuelve un ErrURLInactive (la URL existe pero su dueño la desactivó).
// Ruta EXACTA acordada con el backend: /link-inactivo.
//
// Reusa el patrón visual (ErrorScreen: GridBackground + wordmark + badge con
// ícono), para que se sienta hermana de las páginas de error del propio frontend. Es un
// Server Component público (quien llega aquí solo tenía el link corto, no es un usuario
// logueado). noindex: es una página transitoria de un enlace concreto, no debe indexarse.
export const metadata: Metadata = buildMetadata({
  title: "Enlace no disponible",
  robots: { index: false, follow: false },
});

export default function LinkInactivoPage() {
  return (
    <ErrorScreen
      code={
        <span className="flex size-20 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-accent-500 text-white shadow-lg shadow-brand-600/30">
          <Link2Off className="size-10" aria-hidden />
        </span>
      }
      title="Este enlace ya no está disponible"
      message="El propietario de este enlace lo ha desactivado. ¿Quieres crear y medir tus propios enlaces cortos? Conoce Quikko."
      actions={<CtaLink href="/">Conocer Quikko</CtaLink>}
    />
  );
}
