"use client";

import { ExternalLink } from "lucide-react";
import { Badge, Card } from "@/components/ui";
import { cn } from "@/lib/utils/cn";
import type { ShortURL } from "@/types";
import { QrPanel } from "./QrPanel";

function formatNumber(n: number): string {
  return new Intl.NumberFormat("es").format(n);
}

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("es", {
      day: "numeric",
      month: "long",
      year: "numeric",
    }).format(new Date(iso));
  } catch {
    return "—";
  }
}

function StatusToggle({
  active,
  onToggle,
  disabled,
}: {
  active: boolean;
  onToggle: () => void;
  disabled?: boolean;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={active}
      aria-label={active ? "Desactivar URL" : "Activar URL"}
      onClick={onToggle}
      disabled={disabled}
      className={cn(
        "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white disabled:opacity-50 dark:focus-visible:ring-offset-zinc-900",
        active ? "bg-brand-600" : "bg-zinc-300 dark:bg-zinc-700",
      )}
    >
      <span
        className={cn(
          "inline-block size-4 transform rounded-full bg-white shadow transition-transform",
          active ? "translate-x-4" : "translate-x-0.5",
        )}
      />
    </button>
  );
}

export function UrlDetailHeader({
  url,
  onToggle,
  livePulse,
}: {
  url: ShortURL;
  onToggle: () => void;
  livePulse: boolean;
}) {
  return (
    <Card className="flex flex-col gap-6 sm:flex-row sm:items-start sm:justify-between">
      <div className="min-w-0 flex-1 space-y-4">
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
            /{url.shortCode}
          </h1>
          <Badge variant={url.isActive ? "success" : "neutral"}>
            {url.isActive ? "Activa" : "Inactiva"}
          </Badge>
          {url.isCustomAlias && <Badge variant="brand">Alias</Badge>}
        </div>

        <a
          href={url.shortUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex max-w-full items-center gap-1.5 truncate text-sm font-medium text-brand-600 hover:underline dark:text-brand-400"
        >
          {url.shortUrl}
          <ExternalLink className="size-3.5 shrink-0" aria-hidden />
        </a>

        <dl className="grid gap-4 sm:grid-cols-2">
          <div className="min-w-0">
            <dt className="text-xs font-medium uppercase tracking-wide text-zinc-400">
              Destino
            </dt>
            <dd
              className="mt-0.5 truncate text-sm text-zinc-700 dark:text-zinc-200"
              title={url.originalUrl}
            >
              {url.originalUrl}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-wide text-zinc-400">
              Creada
            </dt>
            <dd className="mt-0.5 text-sm text-zinc-700 dark:text-zinc-200">
              {formatDate(url.createdAt)}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-wide text-zinc-400">
              Clics totales
            </dt>
            <dd className="mt-0.5 flex items-center gap-2 text-sm font-semibold tabular-nums text-zinc-900 dark:text-zinc-50">
              {formatNumber(url.totalClicks)}
              {livePulse && (
                <span className="inline-flex items-center gap-1 rounded-full bg-accent-100 px-1.5 py-0.5 text-[10px] font-semibold text-accent-700 dark:bg-accent-900/40 dark:text-accent-300">
                  <span className="size-1.5 animate-pulse rounded-full bg-accent-500" />
                  nuevo clic
                </span>
              )}
            </dd>
          </div>
          <div>
            <dt className="text-xs font-medium uppercase tracking-wide text-zinc-400">
              Estado
            </dt>
            <dd className="mt-1 flex items-center gap-2">
              <StatusToggle active={url.isActive} onToggle={onToggle} />
              <span className="text-sm text-zinc-600 dark:text-zinc-300">
                {url.isActive ? "Activa" : "Inactiva"}
              </span>
            </dd>
          </div>
        </dl>
      </div>

      {/* QR a un costado en desktop (referencia: panel lateral de Bitly/Dub). */}
      <div className="shrink-0 sm:w-[170px]">
        <QrPanel url={url} />
      </div>
    </Card>
  );
}
