import { cn } from "@/lib/utils/cn";

// Spinner reutilizable. Puramente presentacional. Hereda el color del
// texto (border-current), así se adapta solo a cualquier fondo/tema.
type SpinnerProps = {
  className?: string;
  /** Tamaño en px (alto = ancho). */
  size?: number;
};

export function Spinner({ className, size = 16 }: SpinnerProps) {
  return (
    <span
      role="status"
      aria-label="Cargando"
      style={{ width: size, height: size }}
      className={cn(
        "inline-block shrink-0 animate-spin rounded-full border-2 border-current border-t-transparent",
        className,
      )}
    />
  );
}
