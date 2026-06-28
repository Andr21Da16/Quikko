import { type HTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils/cn";

// Badge. Etiqueta de estado con variantes de color semántico
// (ej. Activa/Inactiva de una URL, Free/Pro del plan). Presentacional.
type BadgeVariant =
  | "neutral"
  | "success"
  | "danger"
  | "warning"
  | "brand"
  | "accent";

const variantClasses: Record<BadgeVariant, string> = {
  neutral:
    "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300",
  success:
    "bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300",
  danger: "bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300",
  warning:
    "bg-amber-100 text-amber-700 dark:bg-amber-900/40 dark:text-amber-300",
  brand:
    "bg-brand-100 text-brand-700 dark:bg-brand-900/40 dark:text-brand-300",
  accent:
    "bg-accent-100 text-accent-800 dark:bg-accent-900/40 dark:text-accent-300",
};

export interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant;
}

export const Badge = forwardRef<HTMLSpanElement, BadgeProps>(
  ({ variant = "neutral", className, ...props }, ref) => (
    <span
      ref={ref}
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  ),
);

Badge.displayName = "Badge";
