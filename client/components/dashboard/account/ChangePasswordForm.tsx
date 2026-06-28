"use client";

import { type FormEvent, useState } from "react";
import { Lock } from "lucide-react";
import { Button, Card, Input } from "@/components/ui";
import { authApi } from "@/lib/api/endpoints/auth";
import { useNotificationsStore } from "@/store/notifications";
import { ApiError } from "@/lib/api/client";

// Cambio de contraseña. El backend exige newPassword de >= 8 chars y la actual
// correcta (devuelve AUTH_INVALID_CREDENTIALS si no coincide, el mismo error de login).
const MIN_PASSWORD = 8;

export function ChangePasswordForm() {
  const notify = useNotificationsStore((s) => s.notify);

  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [currentError, setCurrentError] = useState<string>();
  const [nextError, setNextError] = useState<string>();
  const [confirmError, setConfirmError] = useState<string>();
  const [saving, setSaving] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setCurrentError(undefined);

    const nextInvalid = next.length < MIN_PASSWORD;
    const confirmInvalid = confirm !== next;
    setNextError(
      nextInvalid
        ? `La nueva contraseña debe tener al menos ${MIN_PASSWORD} caracteres.`
        : undefined,
    );
    setConfirmError(confirmInvalid ? "Las contraseñas no coinciden." : undefined);
    if (!current || nextInvalid || confirmInvalid) {
      if (!current) setCurrentError("Introduce tu contraseña actual.");
      return;
    }

    setSaving(true);
    try {
      await authApi.changePassword(current, next);
      notify("success", "Contraseña actualizada.");
      setCurrent("");
      setNext("");
      setConfirm("");
    } catch (err) {
      if (err instanceof ApiError && err.code === "AUTH_INVALID_CREDENTIALS") {
        setCurrentError("La contraseña actual no es correcta.");
      } else {
        notify(
          "error",
          err instanceof ApiError
            ? err.message
            : "No se pudo cambiar la contraseña.",
        );
      }
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-zinc-900 dark:text-zinc-50">
          Contraseña
        </h2>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Cambia tu contraseña. Tu sesión seguirá activa.
        </p>
      </div>

      <Card>
        <form
          onSubmit={handleSubmit}
          className="flex flex-col gap-4 sm:max-w-md"
          noValidate
        >
          <Input
            label="Contraseña actual"
            type="password"
            autoComplete="current-password"
            icon={<Lock size={16} />}
            value={current}
            onChange={(e) => setCurrent(e.target.value)}
            error={currentError}
            required
          />
          <Input
            label="Nueva contraseña"
            type="password"
            autoComplete="new-password"
            placeholder="Mínimo 8 caracteres"
            icon={<Lock size={16} />}
            value={next}
            onChange={(e) => setNext(e.target.value)}
            error={nextError}
            required
          />
          <Input
            label="Confirmar nueva contraseña"
            type="password"
            autoComplete="new-password"
            icon={<Lock size={16} />}
            value={confirm}
            onChange={(e) => setConfirm(e.target.value)}
            error={confirmError}
            required
          />
          <Button type="submit" isLoading={saving} className="self-start">
            Actualizar contraseña
          </Button>
        </form>
      </Card>
    </section>
  );
}
