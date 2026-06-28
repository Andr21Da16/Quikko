// Endpoints del dominio shortener.
import { apiFetch, apiFetchPaginated } from "../client";
import type { ShortURL, CheckAliasResponse, PaginationMeta } from "@/types";

export const urlsApi = {
  createShortURL: (originalUrl: string, customAlias?: string) =>
    apiFetch<ShortURL>("/urls", {
      method: "POST",
      body: JSON.stringify(
        customAlias ? { originalUrl, customAlias } : { originalUrl },
      ),
    }),

  listUserURLs: (
    page = 1,
    limit = 20,
    filters?: { search?: string; isActive?: boolean },
  ): Promise<{ data: ShortURL[]; meta: PaginationMeta | null }> => {
    const params = new URLSearchParams({
      page: String(page),
      limit: String(limit),
    });
    if (filters?.search) params.set("search", filters.search);
    if (filters?.isActive !== undefined) {
      params.set("isActive", String(filters.isActive));
    }
    return apiFetchPaginated<ShortURL[]>(`/urls?${params.toString()}`);
  },

  getURLByCode: (shortCode: string) =>
    apiFetch<ShortURL>(`/urls/code/${encodeURIComponent(shortCode)}`),

  checkAliasAvailability: (alias: string) =>
    apiFetch<CheckAliasResponse>(
      `/urls/check-alias?alias=${encodeURIComponent(alias)}`,
    ),

  toggleActive: (id: string, isActive: boolean) =>
    apiFetch<null>(`/urls/${id}/active`, {
      method: "PATCH",
      body: JSON.stringify({ isActive }),
    }),

  deleteURL: (id: string) =>
    apiFetch<null>(`/urls/${id}`, { method: "DELETE" }),
};
