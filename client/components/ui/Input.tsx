"use client";

import {
  type InputHTMLAttributes,
  type ReactNode,
  forwardRef,
  useId,
} from "react";
import { cn } from "@/lib/utils/cn";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
  /** Icono opcional a la izquierda, ej. <Mail size={16} /> de lucide-react. */
  icon?: ReactNode;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, icon, className, id, ...props }, ref) => {
    const autoId = useId();
    const inputId = id ?? autoId;
    const errorId = error ? `${inputId}-error` : undefined;

    return (
      <div className="flex flex-col gap-1.5">
        {label && (
          <label
            htmlFor={inputId}
            className="text-sm font-medium text-zinc-700 dark:text-zinc-300"
          >
            {label}
          </label>
        )}
        <div className="relative">
          {icon && (
            <span className="pointer-events-none absolute inset-y-0 left-3 flex items-center text-zinc-400 dark:text-zinc-500">
              {icon}
            </span>
          )}
          <input
            ref={ref}
            id={inputId}
            aria-invalid={error ? true : undefined}
            aria-describedby={errorId}
            className={cn(
              "w-full rounded-lg border bg-white px-3 py-2 text-sm text-zinc-900 placeholder:text-zinc-400 transition-colors",
              "focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-brand-500",
              "dark:bg-zinc-900 dark:text-zinc-100 dark:placeholder:text-zinc-500",
              icon && "pl-9",
              error
                ? "border-red-500 focus:ring-red-500 focus:border-red-500"
                : "border-zinc-300 dark:border-zinc-700",
              className,
            )}
            {...props}
          />
        </div>
        {error && (
          <p id={errorId} className="text-sm text-red-600 dark:text-red-400">
            {error}
          </p>
        )}
      </div>
    );
  },
);

Input.displayName = "Input";
