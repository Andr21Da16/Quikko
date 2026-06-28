import { type HTMLAttributes, forwardRef } from "react";
import { cn } from "@/lib/utils/cn";

export const Card = forwardRef<HTMLDivElement, HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => (
    <div
      ref={ref}
      className={cn(
        "rounded-xl border border-zinc-200 bg-white p-6 shadow-sm",
        "dark:border-zinc-800 dark:bg-zinc-900",
        className,
      )}
      {...props}
    />
  ),
);

Card.displayName = "Card";
