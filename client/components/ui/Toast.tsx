"use client";

import { motion } from "framer-motion";
import { CheckCircle, XCircle, Info, X, type LucideIcon } from "lucide-react";
import { cn } from "@/lib/utils/cn";
import type { Notification, NotificationType } from "@/store/notifications";

// Toast individual. Presentacional: recibe la notificación y onDismiss.
// Entrada/salida animadas con Framer Motion (slide + fade + escala, con easing de
// spring). Color/icono semántico; el violeta/lima de marca se reserva para acciones
// primarias. Radio (rounded-xl) consistente con Card. Dark mode completo.
const typeConfig: Record<
  NotificationType,
  { Icon: LucideIcon; accentBorder: string; iconColor: string }
> = {
  success: {
    Icon: CheckCircle,
    accentBorder: "border-l-green-500",
    iconColor: "text-green-600 dark:text-green-400",
  },
  error: {
    Icon: XCircle,
    accentBorder: "border-l-red-500",
    iconColor: "text-red-600 dark:text-red-400",
  },
  info: {
    Icon: Info,
    accentBorder: "border-l-zinc-400 dark:border-l-zinc-500",
    iconColor: "text-zinc-500 dark:text-zinc-400",
  },
};

export function Toast({
  notification,
  onDismiss,
}: {
  notification: Notification;
  onDismiss: (id: string) => void;
}) {
  const { Icon, accentBorder, iconColor } = typeConfig[notification.type];

  return (
    <motion.div
      layout
      initial={{ opacity: 0, x: 32, scale: 0.96 }}
      animate={{ opacity: 1, x: 0, scale: 1 }}
      exit={{ opacity: 0, x: 32, scale: 0.96 }}
      transition={{ type: "spring", stiffness: 380, damping: 30 }}
      role="status"
      aria-live="polite"
      className={cn(
        "pointer-events-auto flex items-start gap-3 rounded-xl border border-l-4 bg-white p-3 pr-2 shadow-lg",
        "border-zinc-200 dark:border-zinc-800 dark:bg-zinc-900",
        accentBorder,
      )}
    >
      <Icon className={cn("mt-0.5 size-5 shrink-0", iconColor)} aria-hidden />
      <p className="flex-1 pt-0.5 text-sm text-zinc-800 dark:text-zinc-100">
        {notification.message}
      </p>
      <button
        type="button"
        onClick={() => onDismiss(notification.id)}
        aria-label="Cerrar notificación"
        className="shrink-0 rounded-md p-1 text-zinc-400 transition-colors hover:bg-zinc-100 hover:text-zinc-600 dark:text-zinc-500 dark:hover:bg-zinc-800 dark:hover:text-zinc-300"
      >
        <X className="size-4" aria-hidden />
      </button>
    </motion.div>
  );
}
