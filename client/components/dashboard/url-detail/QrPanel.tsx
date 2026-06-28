"use client";

/* eslint-disable @next/next/no-img-element */
// El QR es un data URI generado por el backend (no asset remoto ni del filesystem):
// <img> directo es lo correcto; next/image no aporta nada aquí.

import { Check, Copy, Download } from "lucide-react";
import { useClipboard } from "@/lib/hooks/useClipboard";
import type { ShortURL } from "@/types";

export function QrPanel({ url }: { url: ShortURL }) {
  const { copied, copy } = useClipboard();

  // Descarga el data URI como archivo PNG, sin pedir nada al backend (el QR ya es local).
  const download = () => {
    const a = document.createElement("a");
    a.href = url.qrCodeBase64;
    a.download = `quikko-${url.shortCode}.png`;
    document.body.appendChild(a);
    a.click();
    a.remove();
  };

  return (
    <div className="flex flex-col items-center gap-3">
      {url.qrCodeBase64 ? (
        <img
          src={url.qrCodeBase64}
          alt={`Código QR de ${url.shortUrl}`}
          width={150}
          height={150}
          className="size-[150px] rounded-xl border border-zinc-200 bg-white p-2 dark:border-zinc-700"
        />
      ) : (
        <div className="flex size-[150px] items-center justify-center rounded-xl border border-dashed border-zinc-300 text-xs text-zinc-400 dark:border-zinc-700">
          QR no disponible
        </div>
      )}

      <div className="flex w-full gap-2">
        <button
          type="button"
          onClick={() => void copy(url.shortUrl)}
          className="inline-flex h-9 flex-1 items-center justify-center gap-1.5 rounded-lg border border-zinc-300 px-3 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
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
        <button
          type="button"
          onClick={download}
          disabled={!url.qrCodeBase64}
          aria-label="Descargar QR"
          className="inline-flex size-9 items-center justify-center rounded-lg border border-zinc-300 text-zinc-700 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:opacity-40 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
        >
          <Download className="size-4" aria-hidden />
        </button>
      </div>
    </div>
  );
}
