// Endpoints del dominio auth. Cada función usa apiFetch y devuelve tipos
// de types/index.ts. No maneja tokens ni headers: eso es del cliente base.
import { apiFetch } from "../client";
import type { AuthResponse, User, AccountSummary, Plan } from "@/types";

export const authApi = {
  register: (email: string, password: string) =>
    apiFetch<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    apiFetch<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  // El silent refresh lo maneja el cliente base internamente; se expone también aquí
  // por completitud (ej. flujos explícitos).
  refresh: (refreshToken: string) =>
    apiFetch<AuthResponse>("/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refreshToken }),
    }),

  getMe: () => apiFetch<User>("/auth/me"),

  getAccountSummary: () => apiFetch<AccountSummary>("/auth/me/summary"),

  changePassword: (currentPassword: string, newPassword: string) =>
    apiFetch<null>("/auth/me/password", {
      method: "PATCH",
      body: JSON.stringify({ currentPassword, newPassword }),
    }),

  deleteAccount: (password: string) =>
    apiFetch<null>("/auth/me", {
      method: "DELETE",
      body: JSON.stringify({ password }),
    }),

  updatePlan: (plan: Plan) =>
    apiFetch<User>("/auth/me/plan", {
      method: "PATCH",
      body: JSON.stringify({ plan }),
    }),
};
