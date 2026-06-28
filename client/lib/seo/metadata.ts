import type { Metadata } from "next";

// Metadata centralizada de Quikko (Agent 12). buildMetadata define los valores base
// (título/descripción, Open Graph, Twitter, iconos) y permite que cada página los
// sobreescriba parcialmente: `export const metadata = buildMetadata({ title: "Login" })`.
// El título de página (string) hereda el template "%s | Quikko" del layout raíz.
//
// Única fuente de la URL del sitio: NEXT_PUBLIC_SITE_URL (ver client/.env.example).
// Nunca hardcodear otra URL absoluta en otro archivo.

export const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000";
export const SITE_NAME = "Quikko";

const DEFAULT_TITLE =
  "Quikko — Acorta, comparte y mide tus enlaces en tiempo real";
const DEFAULT_DESCRIPTION =
  "Acortador de URLs con analytics en tiempo real: clics por país, dispositivo y navegador, al instante.";

// Imagen OG generada on-demand por el route handler app/og/route.tsx (1200x630).
const OG_IMAGE = "/og";

export function buildMetadata(overrides: Partial<Metadata> = {}): Metadata {
  return {
    metadataBase: new URL(SITE_URL),
    title: { default: DEFAULT_TITLE, template: "%s | Quikko" },
    description: DEFAULT_DESCRIPTION,
    applicationName: SITE_NAME,
    icons: {
      // favicon.ico en app/ lo detecta Next automáticamente; aquí van los PNG.
      icon: [
        { url: "/icons/favicon-16.png", sizes: "16x16", type: "image/png" },
        { url: "/icons/favicon-32.png", sizes: "32x32", type: "image/png" },
        { url: "/icons/favicon-48.png", sizes: "48x48", type: "image/png" },
      ],
      apple: [
        { url: "/icons/favicon-192.png", sizes: "192x192", type: "image/png" },
      ],
    },
    openGraph: {
      type: "website",
      siteName: SITE_NAME,
      title: DEFAULT_TITLE,
      description: DEFAULT_DESCRIPTION,
      url: SITE_URL,
      images: [{ url: OG_IMAGE, width: 1200, height: 630, alt: "Quikko" }],
    },
    twitter: {
      card: "summary_large_image",
      title: DEFAULT_TITLE,
      description: DEFAULT_DESCRIPTION,
      images: [OG_IMAGE],
    },
    ...overrides,
  };
}
