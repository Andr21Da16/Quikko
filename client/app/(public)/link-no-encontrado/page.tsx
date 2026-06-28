import type { Metadata } from "next";
import { SearchX } from "lucide-react";
import { buildMetadata } from "@/lib/seo/metadata";
import { CtaLink } from "@/components/landing/CtaLink";
import { ErrorScreen } from "@/components/error/ErrorScreen";

// Página pública "link no encontrado". El backend redirige aquí
// (302) cuando GET /:code resuelve un ErrURLNotFound (el código no existe, nunca existió o
// se eliminó). Ruta EXACTA acordada con el backend: /link-no-encontrado.
//
// Hermana de /link-inactivo: mismo patrón visual (ErrorScreen), solo cambian ícono y
// mensaje. Distinta también del app/not-found.tsx: aquel es el 404 de ROUTING
// del propio frontend; este es el destino de un short-link inexistente del backend.
export const metadata: Metadata = buildMetadata({
  title: "Enlace no encontrado",
  robots: { index: false, follow: false },
});

export default function LinkNoEncontradoPage() {
  return (
    <ErrorScreen
      code={
        <span className="flex size-20 items-center justify-center rounded-2xl bg-gradient-to-br from-brand-600 to-accent-500 text-white shadow-lg shadow-brand-600/30">
          <SearchX className="size-10" aria-hidden />
        </span>
      }
      title="Este enlace no existe"
      message="No encontramos ningún enlace con esa dirección. Verifica que esté escrito correctamente. ¿Quieres crear los tuyos? Conoce Quikko."
      actions={<CtaLink href="/">Conocer Quikko</CtaLink>}
    />
  );
}
