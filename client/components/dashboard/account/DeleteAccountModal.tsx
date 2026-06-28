"use client";

import { type FormEvent, useState } from "react";
import { AlertTriangle, Lock } from "lucide-react";
import { Button, Input, Modal } from "@/components/ui";
import { authApi } from "@/lib/api/endpoints/auth";
import { ApiError } from "@/lib/api/client";

// Modal de confirmación FUERTE para eliminar la cuenta. Doble barrera contra
// borrados accidentales: (1) re-ingresar la contraseña actual (el backend la exige en el
// body de DELETE /auth/me) y (2) escribir la palabra "ELIMINAR". El botón final queda
// deshabilitado hasta cumplir ambas. Tras éxito, el padre se encarga de logout + redirect.
const CONFIRM_WORD = "ELIMINAR";

export function DeleteAccountModal({
  open,
  onClose,
  onDeleted,
}: {
  open: boolean;
  onClose: () => void;
  onDeleted: () => void;
}) {
  const [password, setPassword] = useState("");
  const [confirmText, setConfirmText] = useState("");
  const [error, setError] = useState<string>();
  const [deleting, setDeleting] = useState(false);

  const canDelete = password.length > 0 && confirmText === CONFIRM_WORD;

  const reset = () => {
    setPassword("");
    setConfirmText("");
    setError(undefined);
  };

  const handleClose = () => {
    if (deleting) return;
    reset();
    onClose();
  };

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (!canDelete) return;
    setError(undefined);
    setDeleting(true);
    try {
      await authApi.deleteAccount(password);
      onDeleted(); // el padre hace logout() + redirect a la landing
    } catch (err) {
      if (err instanceof ApiError && err.code === "AUTH_INVALID_CREDENTIALS") {
        setError("La contraseña no es correcta.");
      } else {
        setError(
          err instanceof ApiError ? err.message : "No se pudo eliminar la cuenta.",
        );
      }
      setDeleting(false);
    }
  }

  return (
    <Modal open={open} onClose={handleClose} title="Eliminar cuenta">
      <form onSubmit={handleSubmit} className="p-6 pt-7" noValidate>
        <div className="flex gap-4">
          <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/40 dark:text-red-400">
            <AlertTriangle className="size-5" aria-hidden />
          </span>
          <div>
            <h2 className="text-base font-semibold text-zinc-900 dark:text-zinc-50">
              Eliminar tu cuenta
            </h2>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              Se eliminarán tu cuenta y todas tus URLs de forma permanente. Esta
              acción no se puede deshacer.
            </p>
          </div>
        </div>

        <div className="mt-5 flex flex-col gap-4">
          <Input
            label="Confirma tu contraseña"
            type="password"
            autoComplete="current-password"
            icon={<Lock size={16} />}
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            error={error}
            required
          />
          <Input
            label={`Escribe "${CONFIRM_WORD}" para confirmar`}
            type="text"
            autoComplete="off"
            value={confirmText}
            onChange={(e) => setConfirmText(e.target.value)}
            required
          />
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <Button
            type="button"
            variant="ghost"
            onClick={handleClose}
            disabled={deleting}
          >
            Cancelar
          </Button>
          <Button
            type="submit"
            variant="danger"
            disabled={!canDelete}
            isLoading={deleting}
          >
            Eliminar mi cuenta
          </Button>
        </div>
      </form>
    </Modal>
  );
}
