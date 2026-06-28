import { type ButtonHTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils/cn";
import { Spinner } from "./Spinner";

type ButtonVariant = "primary" | "secondary" | "ghost" | "danger";
type ButtonSize = "sm" | "md" | "lg";

const base =
  "inline-flex items-center justify-center rounded-lg font-medium transition-colors " +
  "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 " +
  "focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-950 " +
  "disabled:cursor-not-allowed disabled:opacity-50";

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-brand-600 text-white hover:bg-brand-700 focus-visible:ring-brand-500",
  secondary:
    "border border-brand-600 text-brand-700 hover:bg-brand-50 " +
    "dark:border-brand-400 dark:text-brand-300 dark:hover:bg-brand-900/30 " +
    "focus-visible:ring-brand-500",
  ghost:
    "text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800 " +
    "focus-visible:ring-zinc-400",
  danger:
    "bg-red-600 text-white hover:bg-red-700 focus-visible:ring-red-500",
};

const sizeClasses: Record<ButtonSize, string> = {
  sm: "h-8 gap-1.5 px-3 text-sm",
  md: "h-10 gap-2 px-4 text-sm",
  lg: "h-12 gap-2 px-6 text-base",
};

export interface ButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  isLoading?: boolean;
}

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      isLoading = false,
      className,
      children,
      disabled,
      ...props
    },
    ref,
  ) => {
    return (
      <button
        ref={ref}
        // Mientras carga, el botón queda deshabilitado: el navegador no dispara
        // onClick sobre un <button disabled>, cumpliendo el criterio de aceptación.
        disabled={disabled || isLoading}
        aria-busy={isLoading || undefined}
        className={cn(base, variantClasses[variant], sizeClasses[size], className)}
        {...props}
      >
        {isLoading && <Spinner size={16} />}
        {children}
      </button>
    );
  },
);

Button.displayName = "Button";
