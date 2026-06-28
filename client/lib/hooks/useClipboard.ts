"use client";

import { useCallback, useEffect, useRef, useState } from "react";

// Hook de copiado al portapapeles. Usa la Clipboard API del navegador y
// expone un flag `copied` que vuelve a false solo tras `resetMs`, para dar feedback
// visual breve ("¡Copiado!"). Centralizado para reutilizar en cualquier botón de copiar
// (modal de éxito, detalle de URL, etc.).
export function useClipboard(resetMs = 1500): {
  copied: boolean;
  copy: (text: string) => Promise<boolean>;
} {
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    // Limpia el timer pendiente si el componente se desmonta.
    return () => {
      if (timer.current) clearTimeout(timer.current);
    };
  }, []);

  const copy = useCallback(
    async (text: string) => {
      try {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        if (timer.current) clearTimeout(timer.current);
        timer.current = setTimeout(() => setCopied(false), resetMs);
        return true;
      } catch {
        return false;
      }
    },
    [resetMs],
  );

  return { copied, copy };
}
