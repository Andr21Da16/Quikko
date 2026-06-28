// Tipografía única de todo el sitio: Space Grotesk (títulos y cuerpo, landing y
// dashboard). Cargada vía next/font/google: autohosteada, sin layout shift y sin
// <link>/@import manuales. Alternativa de respaldo documentada: Sora (no cambiar sin
// dejar constancia del motivo).
import { Space_Grotesk } from "next/font/google";

export const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  weight: ["300", "400", "500", "600", "700"],
  variable: "--font-space-grotesk",
  display: "swap",
});
