"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button, Card } from "@/components/ui";
import { useAuthStore } from "@/store/auth";
import { useNotificationsStore } from "@/store/notifications";
import { DeleteAccountModal } from "./DeleteAccountModal";

// Zona de peligro: eliminar cuenta. Abre el modal de confirmación fuerte;
// tras eliminar, logout() (limpia store + cookie + desconecta el WS vía la suscripción
// realtime→auth) y redirige a la landing "/" (no a /login: ya no hay cuenta).
export function DangerZone() {
  const router = useRouter();
  const logout = useAuthStore((s) => s.logout);
  const notify = useNotificationsStore((s) => s.notify);
  const [open, setOpen] = useState(false);

  const handleDeleted = () => {
    notify("success", "Tu cuenta ha sido eliminada.");
    logout();
    router.push("/");
  };

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-red-600 dark:text-red-400">
          Zona de peligro
        </h2>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Acciones irreversibles sobre tu cuenta.
        </p>
      </div>

      <Card className="flex flex-col gap-4 border-red-200 sm:flex-row sm:items-center sm:justify-between dark:border-red-900/50">
        <div>
          <p className="font-medium text-zinc-900 dark:text-zinc-50">
            Eliminar cuenta
          </p>
          <p className="mt-0.5 text-sm text-zinc-500 dark:text-zinc-400">
            Borra tu cuenta y todas tus URLs de forma permanente.
          </p>
        </div>
        <Button
          variant="danger"
          onClick={() => setOpen(true)}
          className="shrink-0"
        >
          Eliminar cuenta
        </Button>
      </Card>

      <DeleteAccountModal
        open={open}
        onClose={() => setOpen(false)}
        onDeleted={handleDeleted}
      />
    </section>
  );
}
