"use client";


import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { Input } from "@/components/ui";
import { cn } from "@/lib/utils/cn";
import type { StatusFilter } from "@/store/urls";

const SEARCH_DEBOUNCE_MS = 400;

const STATUS_TABS: { value: StatusFilter; label: string }[] = [
  { value: "all", label: "Todas" },
  { value: "active", label: "Activas" },
  { value: "inactive", label: "Inactivas" },
];

export function UrlsFilters({
  searchQuery,
  statusFilter,
  onSearchChange,
  onStatusChange,
}: {
  searchQuery: string;
  statusFilter: StatusFilter;
  onSearchChange: (q: string) => void;
  onStatusChange: (f: StatusFilter) => void;
}) {
  // Valor inmediato del input (para que teclear sea fluido). Se sincroniza con el
  // término aplicado cuando este cambia desde fuera (ej. "Limpiar filtros") usando el
  // patrón oficial de React "ajustar estado al cambiar una prop" (durante el render, no
  // en un effect — evita la regla react-hooks/set-state-in-effect).
  const [value, setValue] = useState(searchQuery);
  const [lastApplied, setLastApplied] = useState(searchQuery);
  if (searchQuery !== lastApplied) {
    setLastApplied(searchQuery);
    setValue(searchQuery);
  }

  // Debounce: tras una pausa sin teclear, propaga el término. El setState ocurre dentro
  // del callback del timeout (no en el cuerpo del effect). El guard evita re-emitir un
  // valor ya aplicado (incluido el caso de sincronización externa).
  useEffect(() => {
    const trimmed = value.trim();
    if (trimmed === searchQuery) return;
    const t = setTimeout(() => onSearchChange(trimmed), SEARCH_DEBOUNCE_MS);
    return () => clearTimeout(t);
  }, [value, searchQuery, onSearchChange]);

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="sm:max-w-xs sm:flex-1">
        <Input
          type="search"
          placeholder="Buscar por código o URL…"
          aria-label="Buscar URLs"
          icon={<Search className="size-4" aria-hidden />}
          value={value}
          onChange={(e) => setValue(e.target.value)}
        />
      </div>

      <div
        role="tablist"
        aria-label="Filtrar por estado"
        className="inline-flex shrink-0 rounded-lg border border-zinc-200 bg-zinc-50 p-0.5 dark:border-zinc-800 dark:bg-zinc-900"
      >
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            type="button"
            role="tab"
            aria-selected={statusFilter === tab.value}
            onClick={() => onStatusChange(tab.value)}
            className={cn(
              "rounded-md px-3 py-1.5 text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-500",
              statusFilter === tab.value
                ? "bg-white text-brand-700 shadow-sm dark:bg-zinc-800 dark:text-brand-300"
                : "text-zinc-500 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-200",
            )}
          >
            {tab.label}
          </button>
        ))}
      </div>
    </div>
  );
}
