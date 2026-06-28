// Tipos de conveniencia del frontend. Re-exportan con nombres limpios los
// schemas autogenerados desde server/docs/openapi.yaml. El resto del
// frontend importa SIEMPRE desde aquí, NUNCA desde api.generated.ts (autogenerado,
// no se edita a mano: se regenera con `pnpm run generate:types`).
export type { components } from "./api.generated";
import type { components } from "./api.generated";

type Schemas = components["schemas"];

// --- Modelos de dominio (respuestas) ---
export type User = Schemas["UserDTO"];
export type Plan = User["plan"]; // "free" | "pro"
export type ShortURL = Schemas["URLResponse"];
export type ClickStats = Schemas["ClickStats"];
export type TimeBucket = Schemas["TimeBucket"];
export type AccountSummary = Schemas["AccountSummaryResponse"];
export type AuthResponse = Schemas["AuthResponse"];
export type StatsResponse = Schemas["StatsResponse"];
export type CheckAliasResponse = Schemas["CheckAliasResponse"];
export type TimeRange = StatsResponse["range"]; // "24h" | "7d" | "30d"

// --- Cuerpos de petición ---
export type RegisterRequest = Schemas["RegisterRequest"];
export type LoginRequest = Schemas["LoginRequest"];
export type CreateURLRequest = Schemas["CreateURLRequest"];
export type ChangePasswordRequest = Schemas["ChangePasswordRequest"];
export type DeleteAccountRequest = Schemas["DeleteAccountRequest"];
export type UpdatePlanRequest = Schemas["UpdatePlanRequest"];

// --- Realtime (WebSocket) ---
// El WebSocket NO se modela en OpenAPI (ver server/docs/realtime-protocol.md), así
// que estos tipos se mantienen a mano, alineados con el payload `click_event` del
// backend (internal/realtime). Si el protocolo cambia, actualizarlos a mano.
export type ClickEvent = {
  shortCode: string;
  country: string;
  deviceType: string;
  browser: string;
  timestamp: string; // ISO 8601
};

// --- Contrato común (envelope) ---
// El Envelope generado tipa `data: unknown`; aquí lo hacemos genérico y reflejamos
// que data/error pueden ser null según success (sección 4 del documento maestro).
export type ApiErrorBody = Schemas["ApiError"];

export type PaginationMeta = {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type Envelope<T> = {
  success: boolean;
  data: T | null;
  error: ApiErrorBody | null;
  meta?: PaginationMeta;
};
