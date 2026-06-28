import type { Metadata } from "next";
import "./globals.css";
import { spaceGrotesk } from "./fonts";
import { ThemeProvider } from "@/components/theme-provider";
import { ToastContainer } from "@/components/ui/ToastContainer";
import { buildMetadata } from "@/lib/seo/metadata";

// Metadata base centralizada: título + template, descripción, Open Graph,
// Twitter Card e iconos. Cada página puede sobreescribir vía buildMetadata({...}).
export const metadata: Metadata = buildMetadata();

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    // suppressHydrationWarning: next-themes inyecta la clase de tema antes de la
    // hidratación, lo que cambia el <html> respecto al render del server.
    <html
      lang="es"
      suppressHydrationWarning
      className={`${spaceGrotesk.variable} h-full scroll-smooth antialiased`}
    >
      <body className="min-h-full flex flex-col">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
          {children}
          {/* Pila de toasts montada una sola vez para toda la app (Agent 14). */}
          <ToastContainer />
        </ThemeProvider>
      </body>
    </html>
  );
}
