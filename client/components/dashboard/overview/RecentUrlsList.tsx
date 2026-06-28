import Link from "next/link";
import { ArrowRight, Link2, Plus } from "lucide-react";
import { Badge, Card, Spinner } from "@/components/ui";
import type { ShortURL } from "@/types";
import { cn } from "@/lib/utils/cn";

// Lista corta de URLs recientes (Agent 19). Vistazo, no la tabla completa paginada (esa
// es Mis URLs. Cada fila muestra un micro-indicador "+N" cuando llegan clics
// en vivo por WebSocket para esa URL (lo calcula la página padre y se pasa por prop).

function formatDate(iso: string): string {
  try {
    return new Intl.DateTimeFormat("es", {
      day: "numeric",
      month: "short",
    }).format(new Date(iso));
  } catch {
    return "";
  }
}

// Estado vacío amigable: el usuario aún no creó ninguna URL.
function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-3 px-6 py-12 text-center">
      <span className="flex size-12 items-center justify-center rounded-full bg-brand-50 text-brand-600 dark:bg-brand-900/40 dark:text-brand-300">
        <Link2 className="size-6" aria-hidden />
      </span>
      <div>
        <p className="font-medium text-zinc-900 dark:text-zinc-50">
          Aún no tienes ninguna URL
        </p>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Crea tu primer enlace corto para empezar a medir clics.
        </p>
      </div>
      <Link
        href="/dashboard/urls"
        className="inline-flex h-10 items-center gap-2 rounded-lg bg-brand-600 px-4 text-sm font-medium text-white transition-colors hover:bg-brand-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500 focus-visible:ring-offset-2 focus-visible:ring-offset-white dark:focus-visible:ring-offset-zinc-900"
      >
        <Plus className="size-4" aria-hidden />
        Crear mi primera URL
      </Link>
    </div>
  );
}

function UrlRow({ url, liveClicks }: { url: ShortURL; liveClicks: number }) {
  return (
    <li className="flex items-center gap-3 px-5 py-3">
      <div className="min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <span className="truncate text-sm font-medium text-brand-600 dark:text-brand-400">
            /{url.shortCode}
          </span>
          {liveClicks > 0 && (
            <span
              className="inline-flex items-center gap-1 rounded-full bg-accent-100 px-1.5 py-0.5 text-[10px] font-semibold text-accent-700 dark:bg-accent-900/40 dark:text-accent-300"
              title="Clics nuevos en tiempo real"
            >
              <span className="size-1.5 animate-pulse rounded-full bg-accent-500" />
              +{liveClicks}
            </span>
          )}
        </div>
        <p className="truncate text-xs text-zinc-500 dark:text-zinc-400">
          {url.originalUrl}
        </p>
      </div>
      <span className="hidden shrink-0 text-xs text-zinc-400 sm:inline">
        {formatDate(url.createdAt)}
      </span>
      <Badge variant={url.isActive ? "success" : "neutral"}>
        {url.isActive ? "Activa" : "Inactiva"}
      </Badge>
    </li>
  );
}

export function RecentUrlsList({
  urls,
  liveByCode,
  isLoading,
}: {
  urls: ShortURL[];
  liveByCode: Record<string, number>;
  isLoading: boolean;
}) {
  const hasUrls = urls.length > 0;

  return (
    <Card className="p-0">
      <div className="flex items-center justify-between border-b border-zinc-200 px-5 py-3 dark:border-zinc-800">
        <h2 className="text-sm font-semibold text-zinc-900 dark:text-zinc-50">
          URLs recientes
        </h2>
        {hasUrls && (
          <Link
            href="/dashboard/urls"
            className="inline-flex items-center gap-1 text-sm font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
          >
            Ver todas
            <ArrowRight className="size-3.5" aria-hidden />
          </Link>
        )}
      </div>

      {isLoading && !hasUrls ? (
        <div className="flex justify-center py-12">
          <Spinner size={24} className="text-brand-600 dark:text-brand-400" />
        </div>
      ) : hasUrls ? (
        <ul
          className={cn(
            "divide-y divide-zinc-100 dark:divide-zinc-800/80",
            isLoading && "opacity-60",
          )}
        >
          {urls.map((url) => (
            <UrlRow
              key={url.id}
              url={url}
              liveClicks={liveByCode[url.shortCode] ?? 0}
            />
          ))}
        </ul>
      ) : (
        <EmptyState />
      )}
    </Card>
  );
}
