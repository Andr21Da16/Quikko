// Cliente HTTP base. Única capa que sabe de fetch, headers, JWT y del
// envelope { success, data, error, meta }. Los componentes nunca llaman fetch directo.
import type { Envelope, PaginationMeta } from "@/types";
import { session } from "./session";

const API_URL =
  process.env.NEXT_PUBLIC_API_URL;

console.log("API_URL:", API_URL);
// Código del backend que indica access token expirado → dispara el silent refresh.
export const AUTH_TOKEN_EXPIRED = "AUTH_TOKEN_EXPIRED";

// ApiError lleva el `code` estable del backend y el status HTTP, para que el caller
// haga catch y distinga "necesito re-login" de un error de negocio (ej. ALIAS_TAKEN)
// sin volver a parsear nada.
export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "ApiError";
    this.code = code;
    this.status = status;
  }

  // ¿El error obliga a re-autenticarse? (token inválido/expirado y refresh agotado)
  get isAuthError(): boolean {
    return this.status === 401;
  }
}

// --- Silent refresh con single-flight ---
// Si varias llamadas fallan casi a la vez por token expirado, comparten la MISMA
// promesa de refresh en curso en vez de disparar múltiples POST /auth/refresh.
let refreshInFlight: Promise<string | null> | null = null;

async function tryRefreshToken(): Promise<string | null> {
  if (refreshInFlight) return refreshInFlight;

  const refreshToken = session.getRefreshToken();
  if (!refreshToken) return null;

  refreshInFlight = (async () => {
    try {
      // Llamada cruda (no apiFetch) para no recursar en el manejo de 401.
      const res = await fetch(`${API_URL}/auth/refresh`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ refreshToken }),
      });
      const env = (await res.json()) as Envelope<{ accessToken: string }>;
      if (!res.ok || !env.success || !env.data?.accessToken) {
        session.clear();
        return null;
      }
      // El refresh solo emite un nuevo access token; el refresh token no cambia.
      session.setTokens(env.data.accessToken, refreshToken);
      return env.data.accessToken;
    } catch {
      session.clear();
      return null;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

async function requestEnvelope<T>(
  path: string,
  options: RequestInit,
  isRetry: boolean,
): Promise<Envelope<T>> {
  const headers = new Headers(options.headers);
  if (options.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  const token = session.getAccessToken();
  if (token) headers.set("Authorization", `Bearer ${token}`);

  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, { ...options, headers });
  } catch {
    throw new ApiError(
      "NETWORK_ERROR",
      "No se pudo conectar con el servidor.",
      0,
    );
  }

  let env: Envelope<T>;
  try {
    env = (await res.json()) as Envelope<T>;
  } catch {
    throw new ApiError(
      "INVALID_RESPONSE",
      "Respuesta inválida del servidor.",
      res.status,
    );
  }

  if (!env.success || env.error) {
    const code = env.error?.code ?? "UNKNOWN_ERROR";

    // Access token expirado: intentar refresh UNA vez y reintentar la petición.
    if (res.status === 401 && code === AUTH_TOKEN_EXPIRED && !isRetry) {
      const newToken = await tryRefreshToken();
      if (newToken) return requestEnvelope<T>(path, options, true);
      // Si el refresh también falló, tryRefreshToken ya limpió la sesión; el caller
      // recibe el ApiError 401 de abajo y la UI puede redirigir a /login.
    }

    throw new ApiError(code, env.error?.message ?? "Error desconocido.", res.status);
  }

  return env;
}

// apiFetch: para respuestas simples (data: T). Lanza ApiError si success=false.
export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const env = await requestEnvelope<T>(path, options, false);
  return env.data as T;
}

// apiFetchPaginated: para listados; devuelve data + el bloque meta de paginación.
export async function apiFetchPaginated<T>(
  path: string,
  options: RequestInit = {},
): Promise<{ data: T; meta: PaginationMeta | null }> {
  const env = await requestEnvelope<T>(path, options, false);
  return { data: env.data as T, meta: env.meta ?? null };
}

export { API_URL };
