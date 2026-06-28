// Store de notificaciones. Sistema de toasts efímeros con diseño propio
// de Quikko (no una librería genérica). Cada store de dominio decide cuándo llamar
// notify(...) tras una acción; el cliente HTTP NO dispara toasts automáticamente.
import { create } from "zustand";

export type NotificationType = "success" | "error" | "info";

export type Notification = {
  id: string;
  type: NotificationType;
  message: string;
};

// Tiempo de auto-remoción del toast (también se puede cerrar manualmente antes).
const AUTO_DISMISS_MS = 4000;

type NotificationsState = {
  notifications: Notification[];
  // notify agrega un toast (id único) y programa su auto-remoción. Devuelve el id.
  notify: (type: NotificationType, message: string) => string;
  // dismiss elimina un toast manualmente (botón de cerrar) o por timeout.
  dismiss: (id: string) => void;
};

function uid(): string {
  return typeof crypto !== "undefined" && "randomUUID" in crypto
    ? crypto.randomUUID()
    : `${Date.now()}-${Math.random()}`;
}

export const useNotificationsStore = create<NotificationsState>((set, get) => ({
  notifications: [],

  notify: (type, message) => {
    const id = uid();
    set((s) => ({ notifications: [...s.notifications, { id, type, message }] }));
    if (typeof window !== "undefined") {
      window.setTimeout(() => get().dismiss(id), AUTO_DISMISS_MS);
    }
    return id;
  },

  dismiss: (id) =>
    set((s) => ({ notifications: s.notifications.filter((n) => n.id !== id) })),
}));
