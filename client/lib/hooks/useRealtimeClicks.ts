"use client";

import { useMemo } from "react";
import { useRealtimeStore } from "@/store/realtime";
import type { ClickEvent } from "@/types";

// Hook de conveniencia: expone los click_event recientes del store de
// realtime, opcionalmente filtrados por un shortCode (para la pantalla de detalle de
// una URL), evitando que cada componente filtre el array a mano.
export function useRealtimeClicks(shortCode?: string): ClickEvent[] {
  const lastEvents = useRealtimeStore((s) => s.lastEvents);
  return useMemo(
    () =>
      shortCode
        ? lastEvents.filter((e) => e.shortCode === shortCode)
        : lastEvents,
    [lastEvents, shortCode],
  );
}
