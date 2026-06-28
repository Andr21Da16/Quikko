"use client";

// Mis URLs — gestión central: crear, listar, activar/desactivar y eliminar.
// El modal de éxito con QR vive aquí. La gestión completa NO se duplica en Overview.

import { useCallback, useEffect, useState } from "react";
import { authApi } from "@/lib/api/endpoints/auth";
import { useUrlsStore } from "@/store/urls";
import { useNotificationsStore } from "@/store/notifications";
import type { AccountSummary, ShortURL } from "@/types";
import { CreateUrlForm } from "@/components/dashboard/urls/CreateUrlForm";
import { CreatedUrlModal } from "@/components/dashboard/urls/CreatedUrlModal";
import { UrlsTable } from "@/components/dashboard/urls/UrlsTable";
import { UrlsFilters } from "@/components/dashboard/urls/UrlsFilters";
import { DeleteUrlDialog } from "@/components/dashboard/urls/DeleteUrlDialog";
import { PlanQuotaBanner } from "@/components/dashboard/urls/PlanQuotaBanner";
import type { StatusFilter } from "@/store/urls";

export default function UrlsPage() {
  const urls = useUrlsStore((s) => s.urls);
  const pagination = useUrlsStore((s) => s.pagination);
  const isLoading = useUrlsStore((s) => s.isLoading);
  const fetchUrls = useUrlsStore((s) => s.fetchUrls);
  const deleteUrl = useUrlsStore((s) => s.deleteUrl);
  const searchQuery = useUrlsStore((s) => s.searchQuery);
  const statusFilter = useUrlsStore((s) => s.statusFilter);
  const setSearchQuery = useUrlsStore((s) => s.setSearchQuery);
  const setStatusFilter = useUrlsStore((s) => s.setStatusFilter);
  const notify = useNotificationsStore((s) => s.notify);

  const isFiltered = searchQuery !== "" || statusFilter !== "all";

  const [page, setPage] = useState(1);
  const [error, setError] = useState(false);

  // Resumen de cuenta: alimenta el banner de cupo del plan. Vía authApi (capa de datos
  // sancionada).
  const [summary, setSummary] = useState<AccountSummary | null>(null);

  const [createdUrl, setCreatedUrl] = useState<ShortURL | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<ShortURL | null>(null);

  // setState solo dentro de los callbacks de la promesa (no síncrono en el cuerpo), para
  // no disparar react-hooks/set-state-in-effect al llamarse desde el efecto de montaje.
  const loadPage = useCallback(
    (p: number) =>
      fetchUrls(p)
        .then(() => {
          setPage(p);
          setError(false);
        })
        .catch(() => setError(true)),
    [fetchUrls],
  );

  const refreshSummary = useCallback(() => {
    authApi
      .getAccountSummary()
      .then(setSummary)
      .catch(() => setSummary(null));
  }, []);

  // Cambiar un filtro re-dispara el fetch desde la página 1 (los resultados anteriores
  // podrían no existir en la nueva consulta). El setter del store es síncrono, así que
  // loadPage(1) ya lee el filtro actualizado.
  const handleSearchChange = useCallback(
    (q: string) => {
      setSearchQuery(q);
      void loadPage(1);
    },
    [setSearchQuery, loadPage],
  );

  const handleStatusChange = useCallback(
    (f: StatusFilter) => {
      setStatusFilter(f);
      void loadPage(1);
    },
    [setStatusFilter, loadPage],
  );

  const handleClearFilters = useCallback(() => {
    setSearchQuery("");
    setStatusFilter("all");
    void loadPage(1);
  }, [setSearchQuery, setStatusFilter, loadPage]);

  useEffect(() => {
    void loadPage(1);
    refreshSummary();
  }, [loadPage, refreshSummary]);

  // Tras crear: abrir modal de éxito y reconciliar lista (página 1) + cupo.
  const handleCreated = (url: ShortURL) => {
    setCreatedUrl(url);
    void loadPage(1);
    refreshSummary();
  };

  // Borrado confirmado: el store hace update optimista y revierte si falla.
  const handleConfirmDelete = async (url: ShortURL) => {
    try {
      await deleteUrl(url.id);
      notify("success", "URL eliminada.");
      refreshSummary();
      // Si la página actual quedó vacía (y no es la primera), retrocede una.
      if (urls.length === 1 && page > 1) {
        void loadPage(page - 1);
      }
    } catch {
      notify("error", "No se pudo eliminar la URL.");
    }
  };

  return (
    <div className="mx-auto max-w-6xl space-y-6">
      <header>
        <h1 className="text-2xl font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          Mis URLs
        </h1>
        <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
          Crea, gestiona y mide tus enlaces cortos.
        </p>
      </header>

      <CreateUrlForm onCreated={handleCreated} />

      <PlanQuotaBanner summary={summary} />

      <UrlsFilters
        searchQuery={searchQuery}
        statusFilter={statusFilter}
        onSearchChange={handleSearchChange}
        onStatusChange={handleStatusChange}
      />

      <UrlsTable
        urls={urls}
        pagination={pagination}
        isLoading={isLoading}
        error={error}
        isFiltered={isFiltered}
        onRetry={() => void loadPage(page)}
        onPageChange={(p) => void loadPage(p)}
        onRequestDelete={setDeleteTarget}
        onClearFilters={handleClearFilters}
      />

      <CreatedUrlModal url={createdUrl} onClose={() => setCreatedUrl(null)} />

      <DeleteUrlDialog
        url={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleConfirmDelete}
      />
    </div>
  );
}
