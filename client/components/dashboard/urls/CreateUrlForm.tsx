"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { Check, Link2, Loader2, Sparkles, X } from "lucide-react";
import { Button, Card, Input } from "@/components/ui";
import { urlsApi } from "@/lib/api/endpoints/urls";
import { useUrlsStore } from "@/store/urls";
import { useNotificationsStore } from "@/store/notifications";
import { ApiError } from "@/lib/api/client";
import type { ShortURL } from "@/types";

// Formulario de creación. Valida formato de URL en cliente (feedback inmediato;
// el backend es la verdad con ValidateOriginalURL) y verifica disponibilidad de alias en
// tiempo real con debounce. Al crear, delega en el store y avisa al padre para abrir el
// modal de éxito.

// El backend exige alias alfanumérico de 3 a 30 chars (CreateURLRequest validate).
const ALIAS_RE = /^[a-zA-Z0-9]{3,30}$/;
const ALIAS_DEBOUNCE_MS = 450;

// Estado del alias DERIVADO en render (no en un effect), para no caer en
// react-hooks/set-state-in-effect: el único setState async es el resultado del chequeo.
type AliasStatus = "idle" | "invalid" | "checking" | "available" | "taken";

// Resultado de disponibilidad para un valor concreto de alias (evita re-chequear y
// permite saber si el resultado en mano corresponde al texto actual).
type AliasAvailability = { value: string; available: boolean } | null;

// Normaliza la URL: si el usuario omite el esquema, anteponemos https:// (el backend
// rechaza esquemas no http/https, así que no inventamos otros).
function normalizeUrl(raw: string): string {
  const v = raw.trim();
  if (!v) return v;
  return /^https?:\/\//i.test(v) ? v : `https://${v}`;
}

function isLikelyUrl(raw: string): boolean {
  try {
    const u = new URL(normalizeUrl(raw));
    return (u.protocol === "http:" || u.protocol === "https:") && !!u.hostname;
  } catch {
    return false;
  }
}

export function CreateUrlForm({
  onCreated,
}: {
  onCreated: (url: ShortURL) => void;
}) {
  const createUrl = useUrlsStore((s) => s.createUrl);
  const notify = useNotificationsStore((s) => s.notify);

  const [originalUrl, setOriginalUrl] = useState("");
  const [alias, setAlias] = useState("");
  const [urlError, setUrlError] = useState<string>();
  const [availability, setAvailability] = useState<AliasAvailability>(null);
  const [submitting, setSubmitting] = useState(false);
  const [planLimit, setPlanLimit] = useState(false);

  const trimmedAlias = alias.trim();
  const formatValid = ALIAS_RE.test(trimmedAlias);
  const hasResult = availability?.value === trimmedAlias;

  // Estado mostrado, derivado puramente del input + el último resultado conocido.
  const aliasStatus: AliasStatus = !trimmedAlias
    ? "idle"
    : !formatValid
      ? "invalid"
      : hasResult
        ? availability.available
          ? "available"
          : "taken"
        : "checking";

  // Verificación con debounce. setState SOLO dentro del callback async (no en el cuerpo
  // del effect), y un token descarta respuestas obsoletas al teclear rápido.
  const checkToken = useRef(0);
  useEffect(() => {
    if (!trimmedAlias || !ALIAS_RE.test(trimmedAlias)) return;
    if (availability?.value === trimmedAlias) return; // ya tenemos el resultado

    const token = ++checkToken.current;
    const t = setTimeout(async () => {
      try {
        const res = await urlsApi.checkAliasAvailability(trimmedAlias);
        if (token !== checkToken.current) return;
        setAvailability({ value: trimmedAlias, available: res.available });
      } catch {
        if (token !== checkToken.current) return;
        // Fail-open: ante fallo del chequeo no bloqueamos el submit (el backend valida).
        setAvailability({ value: trimmedAlias, available: true });
      }
    }, ALIAS_DEBOUNCE_MS);

    return () => clearTimeout(t);
  }, [trimmedAlias, availability]);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setUrlError(undefined);
    setPlanLimit(false);

    if (!isLikelyUrl(originalUrl)) {
      setUrlError("Introduce una URL válida (ej. https://ejemplo.com/ruta).");
      return;
    }
    if (trimmedAlias && (aliasStatus === "taken" || aliasStatus === "invalid")) {
      return; // el indicador ya muestra el error
    }

    setSubmitting(true);
    try {
      const created = await createUrl(
        normalizeUrl(originalUrl),
        trimmedAlias || undefined,
      );
      // Reset y aviso al padre para abrir el modal de éxito.
      setOriginalUrl("");
      setAlias("");
      setAvailability(null);
      onCreated(created);
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.code) {
          case "INVALID_URL":
            // El backend devuelve el motivo específico (esquema/local/privada/dominio).
            setUrlError(err.message);
            break;
          case "ALIAS_TAKEN":
            setAvailability({ value: trimmedAlias, available: false });
            break;
          case "PLAN_LIMIT_EXCEEDED":
            setPlanLimit(true);
            notify("error", err.message);
            break;
          default:
            notify("error", err.message);
        }
      } else {
        notify("error", "No se pudo crear la URL. Inténtalo de nuevo.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Card>
      <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
        <div className="grid gap-4 sm:grid-cols-[1fr_auto] sm:items-start">
          <div className="flex flex-col gap-4 sm:flex-row">
            <div className="flex-1">
              <Input
                label="URL a acortar"
                type="text"
                inputMode="url"
                placeholder="https://ejemplo.com/una-ruta-muy-larga"
                icon={<Link2 size={16} />}
                value={originalUrl}
                onChange={(e) => setOriginalUrl(e.target.value)}
                error={urlError}
              />
            </div>

            <div className="sm:w-52">
              <Input
                label="Alias (opcional)"
                type="text"
                placeholder="mi-promo"
                value={alias}
                onChange={(e) => setAlias(e.target.value)}
                aria-describedby="alias-status"
              />
              <AliasIndicator status={aliasStatus} />
            </div>
          </div>

          <Button
            type="submit"
            isLoading={submitting}
            className="h-10 w-full sm:mt-[26px] sm:w-auto"
          >
            <Sparkles className="size-4" aria-hidden />
            Acortar
          </Button>
        </div>

        {planLimit && (
          <p className="text-sm text-amber-600 dark:text-amber-400">
            Alcanzaste el límite de URLs activas de tu plan Free.{" "}
            <Link
              href="/dashboard/account"
              className="font-medium underline underline-offset-2 hover:text-amber-700 dark:hover:text-amber-300"
            >
              Mejora a Pro
            </Link>{" "}
            para crear más.
          </p>
        )}
      </form>
    </Card>
  );
}

// Indicador visual de disponibilidad del alias (✓ / ✗ / spinner).
function AliasIndicator({ status }: { status: AliasStatus }) {
  if (status === "idle") return null;

  const map = {
    invalid: {
      icon: <X className="size-3.5" aria-hidden />,
      text: "3-30 caracteres, solo letras y números.",
      cls: "text-red-600 dark:text-red-400",
    },
    checking: {
      icon: <Loader2 className="size-3.5 animate-spin" aria-hidden />,
      text: "Comprobando…",
      cls: "text-zinc-500 dark:text-zinc-400",
    },
    available: {
      icon: <Check className="size-3.5" aria-hidden />,
      text: "Disponible",
      cls: "text-green-600 dark:text-green-400",
    },
    taken: {
      icon: <X className="size-3.5" aria-hidden />,
      text: "No disponible",
      cls: "text-red-600 dark:text-red-400",
    },
  }[status];

  return (
    <p
      id="alias-status"
      className={`mt-1.5 flex items-center gap-1.5 text-xs ${map.cls}`}
    >
      {map.icon}
      {map.text}
    </p>
  );
}
