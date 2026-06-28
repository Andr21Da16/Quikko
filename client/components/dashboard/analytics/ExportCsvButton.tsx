"use client";

import { useState } from "react";
import { Download } from "lucide-react";
import { Button } from "@/components/ui";
import { analyticsApi } from "@/lib/api/endpoints/analytics";
import { useNotificationsStore } from "@/store/notifications";
import type { TimeRange } from "@/types";

// Botón de exportar CSV. Usa analyticsApi.getStatsCSVBlob (fetch autenticado →
// Blob, porque el endpoint exige Authorization y un <a href> nativo no lo envía). La
// descarga se dispara con un <a> temporal + URL.createObjectURL, revocando el object URL.
export function ExportCsvButton({ range }: { range: TimeRange }) {
  const notify = useNotificationsStore((s) => s.notify);
  const [loading, setLoading] = useState(false);

  const handleExport = async () => {
    setLoading(true);
    try {
      const blob = await analyticsApi.getStatsCSVBlob({ range });
      const objectUrl = URL.createObjectURL(blob);
      const a = document.createElement("a");
      a.href = objectUrl;
      a.download = `quikko-analytics-${range}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(objectUrl);
    } catch {
      notify("error", "No se pudo exportar el CSV. Inténtalo de nuevo.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button variant="secondary" onClick={handleExport} isLoading={loading}>
      <Download className="size-4" aria-hidden />
      Exportar CSV
    </Button>
  );
}
