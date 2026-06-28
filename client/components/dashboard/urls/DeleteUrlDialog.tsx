"use client";

import { useState } from "react";
import { AlertTriangle } from "lucide-react";
import { Button, Modal } from "@/components/ui";
import type { ShortURL } from "@/types";


export function DeleteUrlDialog({
  url,
  onClose,
  onConfirm,
}: {
  url: ShortURL | null;
  onClose: () => void;
  onConfirm: (url: ShortURL) => Promise<void>;
}) {
  const [deleting, setDeleting] = useState(false);

  async function handleConfirm() {
    if (!url) return;
    setDeleting(true);
    try {
      await onConfirm(url);
      onClose();
    } finally {
      setDeleting(false);
    }
  }

  return (
    <Modal open={url !== null} onClose={onClose} title="Eliminar URL">
      {url && (
        <div className="p-6 pt-7">
          <div className="flex gap-4">
            <span className="flex size-10 shrink-0 items-center justify-center rounded-full bg-red-100 text-red-600 dark:bg-red-900/40 dark:text-red-400">
              <AlertTriangle className="size-5" aria-hidden />
            </span>
            <div className="min-w-0">
              <h2 className="text-base font-semibold text-zinc-900 dark:text-zinc-50">
                Eliminar esta URL
              </h2>
              <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                Se eliminará{" "}
                <span className="font-medium text-zinc-700 dark:text-zinc-200">
                  /{url.shortCode}
                </span>{" "}
                de forma permanente. Los enlaces que ya compartiste dejarán de
                funcionar. Esta acción no se puede deshacer.
              </p>
            </div>
          </div>

          <div className="mt-6 flex justify-end gap-3">
            <Button variant="ghost" onClick={onClose} disabled={deleting}>
              Cancelar
            </Button>
            <Button variant="danger" onClick={handleConfirm} isLoading={deleting}>
              Eliminar
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
}
