import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";

type UIState = {
  sidebarOpen: boolean;
  toggleSidebar: () => void;
  setSidebarOpen: (open: boolean) => void;
  // Onboarding (Agent 28): el usuario ocultó manualmente la guía de bienvenida. Es una
  // preferencia, no info sensible, así que se persiste en localStorage vía `persist`
  // (mismo patrón que store/auth.ts para tokens, pero solo este campo se persiste; el
  // estado del sidebar es efímero por sesión).
  onboardingDismissed: boolean;
  dismissOnboarding: () => void;
};

export const useUIStore = create<UIState>()(
  persist(
    (set) => ({
      sidebarOpen: false,
      toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),
      setSidebarOpen: (open) => set({ sidebarOpen: open }),
      onboardingDismissed: false,
      dismissOnboarding: () => set({ onboardingDismissed: true }),
    }),
    {
      name: "quikko-ui",
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ onboardingDismissed: state.onboardingDismissed }),
    },
  ),
);
