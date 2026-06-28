import { CtaLink } from "@/components/landing/CtaLink";
import { ErrorScreen } from "@/components/error/ErrorScreen";

// 404 global (convención de Next.js App Router). El root app/not-found.tsx cubre, además
// de las llamadas a notFound(), CUALQUIER ruta del frontend que no existe (ej. alguien
// escribe mal una URL del dashboard) — comportamiento del root not-found desde Next 13.3.
// Es un Server Component sin props; hereda el root layout (ThemeProvider, fuente Space
// Grotesk, estilos), así que dark/light y la tipografía funcionan tal cual. Next inyecta
// automáticamente <meta name="robots" content="noindex"> en respuestas 404.
//
// OJO (distinción de dominio): este 404 es de ROUTING del frontend. No es el caso de un
// shortCode inexistente en el redirect — eso es un error de dominio del backend (GET
// /:code) y nunca llega a este árbol de rutas del cliente.
//
// Salida: un único enlace a "/" es la solución sancionada por la spec (distinguir sesión
// aquí obligaría a leer cookies y volver dinámico el 404, sin beneficio claro: la home ya
// enruta a los usuarios con sesión hacia el dashboard).
export default function NotFound() {
  return (
    <ErrorScreen
      code={
        <span className="bg-gradient-to-br from-brand-600 to-accent-500 bg-clip-text text-7xl font-bold tracking-tight text-transparent sm:text-8xl">
          404
        </span>
      }
      title="Esta página no existe"
      message="La dirección que buscas no está disponible o cambió de lugar. Revisa la URL o vuelve al inicio."
      actions={<CtaLink href="/">Volver al inicio</CtaLink>}
    />
  );
}
