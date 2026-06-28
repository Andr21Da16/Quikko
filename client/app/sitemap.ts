import type { MetadataRoute } from "next";
import { SITE_URL } from "@/lib/seo/metadata";

// Genera /sitemap.xml en build (convención App Router). Solo se listan rutas públicas
// que YA existen: hoy únicamente la home (`/`). /login y /register se agregarán aquí
// cuando el agente que construye la landing las cree como páginas reales — no se
// inventan rutas que aún devolverían 404.
export default function sitemap(): MetadataRoute.Sitemap {
  return [{ url: SITE_URL, lastModified: new Date(), priority: 1 }];
}
