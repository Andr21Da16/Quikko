"use client";

// Wrapper cliente sobre next-themes. El layout raíz es un Server Component, así que
// la conexión con el provider (que usa hooks/efectos de cliente) se aísla aquí.
import { ThemeProvider as NextThemesProvider } from "next-themes";
import type { ComponentProps } from "react";

export function ThemeProvider({
  children,
  ...props
}: ComponentProps<typeof NextThemesProvider>) {
  return <NextThemesProvider {...props}>{children}</NextThemesProvider>;
}
