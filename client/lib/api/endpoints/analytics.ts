// Endpoints del dominio analytics.
import { apiFetch, API_URL, ApiError } from "../client";
import { session } from "../session";
import type { StatsResponse, TimeRange, Envelope } from "@/types";

type StatsParams = { shortCode?: string; range?: TimeRange };

function statsQuery({ shortCode, range }: StatsParams): string {
  const q = new URLSearchParams();
  if (shortCode) q.set("shortCode", shortCode);
  if (range) q.set("range", range);
  const qs = q.toString();
  return qs ? `?${qs}` : "";
}

export const analyticsApi = {
  getStats: (params: StatsParams = {}) =>
    apiFetch<StatsResponse>(`/analytics/stats${statsQuery(params)}`),

  // Export CSV: el endpoint responde text/csv (NO el envelope JSON), así que no pasa
  // por apiFetch<T>. DECISIÓN: se eligió el enfoque "fetch crudo → Blob" (en vez de
  // devolver una URL para <a href>), porque el endpoint está protegido por JWT en el
  // header Authorization; un <a href> nativo no enviaría ese header y daría 401. El
  // Blob permite adjuntar el token y luego descargarlo con un object URL en la UI.
  // Los errores de este endpoint SÍ usan el envelope JSON (lo traducimos a ApiError).
  getStatsCSVBlob: async (params: StatsParams = {}): Promise<Blob> => {
    const headers = new Headers();
    const token = session.getAccessToken();
    if (token) headers.set("Authorization", `Bearer ${token}`);

    const res = await fetch(
      `${API_URL}/analytics/stats/export${statsQuery(params)}`,
      { headers },
    );

    if (!res.ok) {
      let code = "EXPORT_FAILED";
      let message = "No se pudo exportar el CSV.";
      try {
        const env = (await res.json()) as Envelope<unknown>;
        code = env.error?.code ?? code;
        message = env.error?.message ?? message;
      } catch {
        // respuesta no-JSON inesperada: se mantienen los valores por defecto.
      }
      throw new ApiError(code, message, res.status);
    }

    return res.blob();
  },
};
