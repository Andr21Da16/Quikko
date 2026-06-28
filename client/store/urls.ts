import { create } from "zustand";
import { urlsApi } from "@/lib/api/endpoints/urls";
import type { ShortURL, PaginationMeta } from "@/types";

export type StatusFilter = "all" | "active" | "inactive";

const PAGE_SIZE = 20;

type UrlsState = {
  urls: ShortURL[];
  pagination: PaginationMeta | null;
  isLoading: boolean;
  // Filtros activos. El filtrado se delega al backend, no se hace en cliente.
  searchQuery: string;
  statusFilter: StatusFilter;

  fetchUrls: (page?: number) => Promise<void>;
  setSearchQuery: (q: string) => void;
  setStatusFilter: (f: StatusFilter) => void;
  createUrl: (originalUrl: string, customAlias?: string) => Promise<ShortURL>;
  toggleActive: (id: string, isActive: boolean) => Promise<void>;
  deleteUrl: (id: string) => Promise<void>;
};

export const useUrlsStore = create<UrlsState>((set, get) => ({
  urls: [],
  pagination: null,
  isLoading: false,
  searchQuery: "",
  statusFilter: "all",

  fetchUrls: async (page = 1) => {
    set({ isLoading: true });
    try {
      const { searchQuery, statusFilter } = get();
      const { data, meta } = await urlsApi.listUserURLs(page, PAGE_SIZE, {
        search: searchQuery || undefined,
        isActive: statusFilter === "all" ? undefined : statusFilter === "active",
      });
      set({ urls: data, pagination: meta });
    } finally {
      set({ isLoading: false });
    }
  },

  // Los setters solo guardan el filtro; re-disparar el fetch desde la página 1 lo
  // coordina la página (mantiene su estado local de paginación en sync).
  setSearchQuery: (q) => set({ searchQuery: q }),
  setStatusFilter: (f) => set({ statusFilter: f }),

  createUrl: async (originalUrl, customAlias) => {
    const created = await urlsApi.createShortURL(originalUrl, customAlias);
    // Actualiza la lista sin refetch: el nuevo registro va al principio.
    set((s) => ({ urls: [created, ...s.urls] }));
    return created;
  },

  toggleActive: async (id, isActive) => {
    const prev = get().urls;
    // Optimista: refleja el cambio antes de confirmar con el servidor.
    set({ urls: prev.map((u) => (u.id === id ? { ...u, isActive } : u)) });
    try {
      await urlsApi.toggleActive(id, isActive);
    } catch (err) {
      set({ urls: prev }); // revertir
      throw err;
    }
  },

  deleteUrl: async (id) => {
    const prev = get().urls;
    set({ urls: prev.filter((u) => u.id !== id) }); // optimista
    try {
      await urlsApi.deleteURL(id);
    } catch (err) {
      set({ urls: prev }); // revertir
      throw err;
    }
  },
}));
