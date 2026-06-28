"use client";

import { AnimatePresence } from "framer-motion";
import { useNotificationsStore } from "@/store/notifications";
import { Toast } from "./Toast";

// ToastContainer. Se monta UNA sola vez en el layout raíz. Lee el store de
// notificaciones y apila los toasts en la esquina inferior derecha. El contenedor es
// pointer-events-none (no bloquea clics de la página); cada Toast reactiva los suyos.
export function ToastContainer() {
  const notifications = useNotificationsStore((s) => s.notifications);
  const dismiss = useNotificationsStore((s) => s.dismiss);

  return (
    <div className="pointer-events-none fixed bottom-4 right-4 z-50 flex w-full max-w-sm flex-col gap-2">
      <AnimatePresence initial={false}>
        {notifications.map((n) => (
          <Toast key={n.id} notification={n} onDismiss={dismiss} />
        ))}
      </AnimatePresence>
    </div>
  );
}
