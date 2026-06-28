"use client";

import { useTheme } from "next-themes";
import { Moon, Sun } from "lucide-react";
import { cn } from "@/lib/utils/cn";

// Toggle de dark mode. Usa next-themes. En vez del clásico guard `mounted`
// (que provoca setState-en-effect), ambos iconos se renderizan siempre y el CSS muestra
// el correcto según la clase `dark` del <html> (variante dark de Tailwind). Así no hay
// mismatch de hidratación: el markup no depende del tema resuelto en el primer render.
export function ThemeToggle({ className }: { className?: string }) {
  const { resolvedTheme, setTheme } = useTheme();

  return (
    <button
      type="button"
      aria-label="Cambiar tema"
      onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
      className={cn(
        "inline-flex size-9 items-center justify-center rounded-lg text-zinc-600 transition-colors hover:bg-zinc-100 hover:text-zinc-900 dark:text-zinc-300 dark:hover:bg-zinc-800 dark:hover:text-zinc-50",
        className,
      )}
    >
      {/* En claro se ve la luna (ir a oscuro); en oscuro, el sol (ir a claro). */}
      <Moon className="size-5 dark:hidden" aria-hidden />
      <Sun className="hidden size-5 dark:block" aria-hidden />
    </button>
  );
}
