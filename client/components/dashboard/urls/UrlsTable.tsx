"use client";

import Link from "next/link";
import {
  ChevronLeft,
  ChevronRight,
  Link2,
  Plus,
  SearchX,
  Trash2,
} from "lucide-react";
import { Badge, Card, Spinner } from "@/components/ui";
import { useUrlsStore } from "@/store/urls";
import { useNotificationsStore } from "@/store/notifications";
import type { PaginationMeta, ShortURL } from "@/types";
import { cn } from "@/lib/utils/cn";

function formatNumber(n: number): string {
  return new Intl.NumberFormat("es").format(n);
}

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("es", {
      day: "numeric",
      month: "short",
      year: "numeric",
    }).format(new Date(iso));
  } catch {
    return "";
  }
}

// Switch de activar/desactivar. El store ya hace update optimista + revertir si falla;
// aquí solo notificamos el fallo.
function ActiveToggle({ url }: { url: ShortURL }) {
  const toggleActive = useUrlsStore((s) => s.toggleActive);
  const notify = useNotificationsStore((s) => s.notify);

  const handle = async () => {
    try {
      await toggleActive(url.id, !url.isActive);
    } catch {
      notify("error", "No se pudo cambiar el estado de la URL.");
    }
  };

  return (
    <button
      type="button"
      role="switch"
      aria-checked={url.isActive}
      aria-label={url.isActive ? "Desactivar URL" : "Activar URL"}
      onClick={handle}
      className={cn(
        "relative inline-flex h-5 w-9 shrink-0 items-center rounded-full transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900",
        url.isActive ? "bg-brand-600" : "bg-zinc-300 dark:bg-zinc-700",
      )}
    >
      <span
        className={cn(
          "inline-block size-4 transform rounded-full bg-white shadow transition-transform",
          url.isActive ? "translate-x-4" : "translate-x-0.5",
        )}
      />
    </button>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-16 text-center">
      <span className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600 dark:bg-brand-900/40 dark:text-brand-300">
        <Link2 className="size-6" aria-hidden />
      </span>
      <div>
        <p className="font-medium text-zinc-900 dark:text-zinc-50">
          Aún no tienes ninguna URL
        </p>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Usa el formulario de arriba para crear tu primer enlace corto.
        </p>
      </div>
    </div>
  );
}

// Estado vacío específico para búsqueda/filtro sin coincidencias (Agent 27). Distinto
// del onboarding ("aún no tienes URLs"): aquí el usuario sí tiene URLs, solo que ninguna
// matchea el filtro actual.
function NoMatchesState({ onClearFilters }: { onClearFilters: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-16 text-center">
      <span className="flex size-12 items-center justify-center rounded-full bg-zinc-100 text-zinc-500 dark:bg-zinc-800 dark:text-zinc-400">
        <SearchX className="size-6" aria-hidden />
      </span>
      <div>
        <p className="font-medium text-zinc-900 dark:text-zinc-50">
          No encontramos URLs que coincidan con tu búsqueda
        </p>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Prueba con otro término o limpia los filtros.
        </p>
      </div>
      <button
        type="button"
        onClick={onClearFilters}
        className="inline-flex h-9 items-center rounded-lg border border-zinc-300 px-4 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
      >
        Limpiar filtros
      </button>
    </div>
  );
}

function ErrorState({ onRetry }: { onRetry: () => void }) {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-16 text-center">
      <p className="text-sm text-zinc-500 dark:text-zinc-400">
        No se pudieron cargar tus URLs.
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="inline-flex h-9 items-center gap-2 rounded-lg border border-zinc-300 px-4 text-sm font-medium text-zinc-700 transition-colors hover:bg-zinc-100 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
      >
        <Plus className="size-4 rotate-45" aria-hidden />
        Reintentar
      </button>
    </div>
  );
}

export function UrlsTable({
  urls,
  pagination,
  isLoading,
  error,
  isFiltered,
  onRetry,
  onPageChange,
  onRequestDelete,
  onClearFilters,
}: {
  urls: ShortURL[];
  pagination: PaginationMeta | null;
  isLoading: boolean;
  error: boolean;
  /** Hay un filtro/búsqueda activo: cambia el estado vacío (sin coincidencias vs onboarding). */
  isFiltered: boolean;
  onRetry: () => void;
  onPageChange: (page: number) => void;
  onRequestDelete: (url: ShortURL) => void;
  onClearFilters: () => void;
}) {
  const hasUrls = urls.length > 0;

  if (error && !hasUrls) {
    return (
      <Card className="p-0">
        <ErrorState onRetry={onRetry} />
      </Card>
    );
  }

  if (isLoading && !hasUrls) {
    return (
      <Card className="flex justify-center py-16">
        <Spinner size={24} className="text-brand-600 dark:text-brand-400" />
      </Card>
    );
  }

  if (!hasUrls) {
    return (
      <Card className="p-0">
        {isFiltered ? (
          <NoMatchesState onClearFilters={onClearFilters} />
        ) : (
          <EmptyState />
        )}
      </Card>
    );
  }

  const page = pagination?.page ?? 1;
  const totalPages = pagination?.totalPages ?? 1;

  return (
    <Card className="overflow-hidden p-0">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm">
          <thead className="border-b border-zinc-200 text-xs uppercase tracking-wide text-zinc-500 dark:border-zinc-800 dark:text-zinc-400">
            <tr>
              <th className="px-5 py-3 font-medium">Enlace</th>
              <th className="px-3 py-3 text-right font-medium">Clics</th>
              <th className="hidden px-3 py-3 font-medium sm:table-cell">Estado</th>
              <th className="hidden px-3 py-3 font-medium md:table-cell">Creada</th>
              <th className="px-5 py-3 text-right font-medium">Acciones</th>
            </tr>
          </thead>
          <tbody
            className={cn(
              "divide-y divide-zinc-100 dark:divide-zinc-800/80",
              isLoading && "opacity-60",
            )}
          >
            {urls.map((url) => (
              <tr key={url.id} className="align-middle">
                <td className="max-w-0 px-5 py-3">
                  <Link
                    href={`/dashboard/urls/${url.shortCode}`}
                    className="font-medium text-brand-600 hover:underline dark:text-brand-400"
                  >
                    /{url.shortCode}
                  </Link>
                  <p
                    className="truncate text-xs text-zinc-500 dark:text-zinc-400"
                    title={url.originalUrl}
                  >
                    {url.originalUrl}
                  </p>
                </td>
                <td className="px-3 py-3 text-right font-medium tabular-nums text-zinc-700 dark:text-zinc-200">
                  {formatNumber(url.totalClicks)}
                </td>
                <td className="hidden px-3 py-3 sm:table-cell">
                  <Badge variant={url.isActive ? "success" : "neutral"}>
                    {url.isActive ? "Activa" : "Inactiva"}
                  </Badge>
                </td>
                <td className="hidden whitespace-nowrap px-3 py-3 text-zinc-500 dark:text-zinc-400 md:table-cell">
                  {formatDate(url.createdAt)}
                </td>
                <td className="px-5 py-3">
                  <div className="flex items-center justify-end gap-3">
                    <ActiveToggle url={url} />
                    <button
                      type="button"
                      onClick={() => onRequestDelete(url)}
                      aria-label={`Eliminar ${url.shortCode}`}
                      className="inline-flex size-8 items-center justify-center rounded-lg text-zinc-400 transition-colors hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/30 dark:hover:text-red-400"
                    >
                      <Trash2 className="size-4" aria-hidden />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Paginación */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between border-t border-zinc-200 px-5 py-3 text-sm dark:border-zinc-800">
          <span className="text-zinc-500 dark:text-zinc-400">
            Página {page} de {totalPages}
          </span>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => onPageChange(page - 1)}
              disabled={page <= 1 || isLoading}
              className="inline-flex h-8 items-center gap-1 rounded-lg border border-zinc-300 px-2.5 font-medium text-zinc-700 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:opacity-40 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              <ChevronLeft className="size-4" aria-hidden />
              Anterior
            </button>
            <button
              type="button"
              onClick={() => onPageChange(page + 1)}
              disabled={page >= totalPages || isLoading}
              className="inline-flex h-8 items-center gap-1 rounded-lg border border-zinc-300 px-2.5 font-medium text-zinc-700 transition-colors hover:bg-zinc-100 disabled:cursor-not-allowed disabled:opacity-40 dark:border-zinc-700 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              Siguiente
              <ChevronRight className="size-4" aria-hidden />
            </button>
          </div>
        </div>
      )}
    </Card>
  );
}
