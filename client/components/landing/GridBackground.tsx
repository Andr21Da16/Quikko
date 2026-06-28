import { cn } from "@/lib/utils/cn";

// Patrón "blueprint"/engineering grid del hero. Técnica de Vercel/Linear:
// dos linear-gradients (líneas verticales + horizontales) a muy baja opacidad, celdas
// de 64px, desvanecidas hacia los bordes con un mask radial para que no compitan con
// el contenido. Casi subliminal, nunca ruidoso.
//
// Dos capas porque el color de línea cambia por modo: líneas claras sobre fondo oscuro
// (dark) y líneas oscuras sobre fondo claro (light). Son líneas estructurales neutras
// (no colores de marca), por eso van con alfa fija y no con tokens brand/accent.
export function GridBackground({ className }: { className?: string }) {
  const cell = "bg-[length:64px_64px]";
  const mask =
    "[mask-image:radial-gradient(ellipse_70%_60%_at_50%_35%,#000_55%,transparent_100%)]";

  return (
    <div
      aria-hidden
      className={cn("pointer-events-none absolute inset-0 overflow-hidden", className)}
    >
      {/* Modo claro: líneas oscuras tenues. */}
      <div
        className={cn(
          "absolute inset-0 dark:hidden",
          "bg-[linear-gradient(to_right,#0f172a0f_1px,transparent_1px),linear-gradient(to_bottom,#0f172a0f_1px,transparent_1px)]",
          cell,
          mask,
        )}
      />
      {/* Modo oscuro: líneas claras tenues. */}
      <div
        className={cn(
          "absolute inset-0 hidden dark:block",
          "bg-[linear-gradient(to_right,#ffffff0a_1px,transparent_1px),linear-gradient(to_bottom,#ffffff0a_1px,transparent_1px)]",
          cell,
          mask,
        )}
      />
    </div>
  );
}
