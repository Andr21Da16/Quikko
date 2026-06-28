// Store de auth. Única excepción consciente a "no usar localStorage en
// componentes": el store SÍ persiste los tokens vía el middleware `persist`, pero
// ningún componente toca localStorage por su cuenta.
import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import { authApi } from "@/lib/api/endpoints/auth";
import { registerSession } from "@/lib/api/session";
import type { Plan, User } from "@/types";

// Cookie de sesión leída por client/middleware.ts para proteger /dashboard/*.
//
// TRADE-OFF documentado: se escribe desde JS (NO httpOnly — JS no puede setear httpOnly),
// así que NO es el mecanismo de máxima seguridad. Es coherente con que el JWT real ya
// viaja por header Authorization (no por cookie). El middleware solo comprueba que la
// cookie EXISTA (barrera de UX anti-parpadeo); la autorización real la hace el backend
// en cada request. La cookie solo guarda el access token para esa comprobación.
const SESSION_COOKIE = "quikko_session";
// Alineado con la vida del refresh token (7d): cubre la sesión recordada.
const COOKIE_MAX_AGE = 60 * 60 * 24 * 7;

function writeSessionCookie(token: string): void {
  if (typeof document === "undefined") return;
  document.cookie = `${SESSION_COOKIE}=${token}; path=/; max-age=${COOKIE_MAX_AGE}; SameSite=Lax`;
}

function clearSessionCookie(): void {
  if (typeof document === "undefined") return;
  document.cookie = `${SESSION_COOKIE}=; path=/; max-age=0; SameSite=Lax`;
}

type AuthState = {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;

  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string) => Promise<void>;
  logout: () => void;
  // setTokens lo usa el silent refresh del cliente HTTP (vía session).
  setTokens: (accessToken: string, refreshToken: string) => void;
  fetchCurrentUser: () => Promise<void>;
  // updatePlan cambia el plan en el backend y refleja el User devuelto en el store, para
  // que el resto de la app (badge del shell, gating de rango) reaccione sin recargar
  // (el plan NO viaja en el JWT, por eso basta con actualizar el estado local).
  updatePlan: (plan: Plan) => Promise<void>;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
      isLoading: false,

      login: async (email, password) => {
        set({ isLoading: true });
        try {
          const res = await authApi.login(email, password);
          set({
            accessToken: res.accessToken,
            refreshToken: res.refreshToken ?? null,
            user: res.user ?? null,
            isAuthenticated: true,
          });
          writeSessionCookie(res.accessToken);
        } finally {
          set({ isLoading: false });
        }
      },

      register: async (email, password) => {
        set({ isLoading: true });
        try {
          const res = await authApi.register(email, password);
          set({
            accessToken: res.accessToken,
            refreshToken: res.refreshToken ?? null,
            user: res.user ?? null,
            isAuthenticated: true,
          });
          writeSessionCookie(res.accessToken);
        } finally {
          set({ isLoading: false });
        }
      },

      logout: () => {
        set({
          user: null,
          accessToken: null,
          refreshToken: null,
          isAuthenticated: false,
        });
        clearSessionCookie();
      },

      // setTokens lo invoca el silent refresh del cliente HTTP: refrescar el access
      // token también refresca la cookie de sesión.
      setTokens: (accessToken, refreshToken) => {
        set({ accessToken, refreshToken, isAuthenticated: true });
        writeSessionCookie(accessToken);
      },

      fetchCurrentUser: async () => {
        // Tras un F5, el token se rehidrata pero el `user` no se persiste: se vuelve
        // a pedir aquí. Si el token ya no sirve, el cliente HTTP limpia la sesión.
        if (!get().accessToken) return;
        set({ isLoading: true });
        try {
          const user = await authApi.getMe();
          set({ user, isAuthenticated: true });
        } finally {
          set({ isLoading: false });
        }
      },

      updatePlan: async (plan) => {
        const user = await authApi.updatePlan(plan);
        set({ user });
      },
    }),
    {
      name: "quikko-auth",
      storage: createJSONStorage(() => localStorage),
      // Solo se persisten los tokens; el `user` se re-pide con fetchCurrentUser.
      partialize: (state) => ({
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
      }),
      // Tras rehidratar, recomputar isAuthenticated a partir del token restaurado.
      onRehydrateStorage: () => (state) => {
        if (state) state.isAuthenticated = !!state.accessToken;
      },
    },
  ),
);

// El store se registra como fuente de sesión para el cliente HTTP (lee/escribe tokens
// y limpia sesión en refresh fallido), sin que el cliente importe el store. (Agent 13)
registerSession({
  getAccessToken: () => useAuthStore.getState().accessToken,
  getRefreshToken: () => useAuthStore.getState().refreshToken,
  setTokens: (accessToken, refreshToken) =>
    useAuthStore.getState().setTokens(accessToken, refreshToken),
  clear: () => useAuthStore.getState().logout(),
});
