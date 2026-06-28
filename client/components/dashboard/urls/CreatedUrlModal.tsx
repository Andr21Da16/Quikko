"use client";

/* eslint-disable @next/next/no-img-element */
// El QR es un data URI generado por el backend (no un asset remoto ni del filesystem),
// así que next/image no aporta nada aquí: <img> directo es lo correcto.

import Link from "next/link";
import { ArrowUpRight, Check, Copy, PartyPopper } from "lucide-react";
import { Modal } from "@/components/ui";
import { useClipboard } from "@/lib/hooks/useClipboard";
import type { ShortURL } from "@/types";

// Modal de éxito al crear. Investigación (Bitly, Dub.co, TinyURL, Short.io):
// este momento se presenta como un pequeño logro, con el link corto destacado y el copiar
// como acción primaria, el QR visible al lado, y un cierre discreto — no un alert(). El
// QR (qrCodeBase64) ya viene en la respuesta de creación como data URI.
export function CreatedUrlModal({
  url,
  onClose,
}: {
  url: ShortURL | null;
  onClose: () => void;
}) {
  const { copied, copy } = useClipboard();

  return (
    <Modal open={url !== null} onClose={onClose} title="Enlace creado">
      {url && (
        <div className="p-6 pt-7">
          <div className="flex flex-col items-center text-center">
            <span className="flex size-11 items-center justify-center rounded-full bg-accent-100 text-accent-700 dark:bg-accent-900/40 dark:text-accent-300">
              <PartyPopper className="size-5" aria-hidden />
            </span>
            <h2 className="mt-3 text-lg font-bold text-zinc-900 dark:text-zinc-50">
              ¡Tu enlace está listo!
            </h2>
            <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
              Compártelo o escanéalo. Ya estamos midiendo sus clics.
            </p>
          </div>

          {/* QR */}
          {url.qrCodeBase64 ? (
            <div className="mt-5 flex justify-center">
              <img
                src={url.qrCodeBase64}
                alt={`Código QR de ${url.shortUrl}`}
                width={160}
                height={160}
                className="size-40 rounded-xl border border-zinc-200 bg-white p-2 dark:border-zinc-700"
              />
            </div>
          ) : null}

          {/* Link corto + copiar (acción primaria) */}
          <div className="mt-5">
            <label className="text-xs font-medium text-zinc-500 dark:text-zinc-400">
              Tu enlace corto
            </label>
            <div className="mt-1.5 flex items-center gap-2">
              <input
                readOnly
                value={url.shortUrl}
                onFocus={(e) => e.currentTarget.select()}
                className="min-w-0 flex-1 truncate rounded-lg border border-zinc-300 bg-zinc-50 px-3 py-2 text-sm text-zinc-900 dark:border-zinc-700 dark:bg-zinc-800 dark:text-zinc-100"
              />
              <button
                type="button"
                onClick={() => void copy(url.shortUrl)}
                aria-label="Copiar enlace"
                className="inline-flex h-10 shrink-0 items-center gap-2 rounded-lg bg-brand-600 px-3 text-sm font-medium text-white transition-colors hover:bg-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900"
              >
                {copied ? (
                  <>
                    <Check className="size-4" aria-hidden />
                    ¡Copiado!
                  </>
                ) : (
                  <>
                    <Copy className="size-4" aria-hidden />
                    Copiar
                  </>
                )}
              </button>
            </div>
          </div>

          {/* Acciones secundarias */}
          <div className="mt-6 flex items-center justify-between gap-3">
            <Link
              href={`/dashboard/urls/${url.shortCode}`}
              className="inline-flex items-center gap-1 text-sm font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
            >
              Ver detalle
              <ArrowUpRight className="size-4" aria-hidden />
            </Link>
            <button
              type="button"
              onClick={onClose}
              className="inline-flex h-10 items-center rounded-lg border border-zinc-300 px-4 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              Cerrar
            </button>
          </div>
        </div>
      )}
    </Modal>
  );
}
