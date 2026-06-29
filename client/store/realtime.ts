// Store de WebSocket en tiempo real. Conexión ÚNICA y compartida: ninguna
// pantalla abre su propia conexión. El store solo distribuye el dato (lastEvents);
// cada pantalla decide qué hacer visualmente con un click_event.
//
// Dependencias en una sola dirección: realtime → auth (lee el token) y
// realtime → notifications (traduce errores del server a toasts). auth NO importa
// realtime; en su lugar, este store SE SUSCRIBE a auth (abajo) para conectar al
// iniciar sesión y desconectar al cerrarla — equivalente a llamar connect()/disconnect()
// desde login()/logout(), pero sin crear un ciclo de imports.
import { create } from "zustand";
import type { ClickEvent } from "@/types";
import { useAuthStore } from "./auth";
import { useNotificationsStore } from "./notifications";

const WS_URL = process.env.NEXT_PUBLIC_WS_URL;
const MAX_EVENTS = 50; // buffer corto de contexto reciente
const BASE_BACKOFF_MS = 1000;
const MAX_BACKOFF_MS = 30000;

// Forma de los mensajes del protocolo (server/docs/realtime-protocol.md).
type ClickEventPayload = ClickEvent;
type ErrorPayload = { code: string; message: string };
type ServerMessage =
  | { type: "click_event"; payload: ClickEventPayload }
  | { type: "error"; payload: ErrorPayload };

type RealtimeState = {
  socket: WebSocket | null;
  isConnected: boolean;
  lastEvents: ClickEvent[];
  connect: () => void;
  disconnect: () => void;
  subscribeToUrl: (shortCode: string) => void;
  unsubscribeFromUrl: (shortCode: string) => void;
};

// Estado de reconexión fuera del store (hay una sola conexión global).
let reconnectAttempts = 0;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let manualClose = false;

export const useRealtimeStore = create<RealtimeState>((set, get) => {
  function clearReconnectTimer() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer);
      reconnectTimer = null;
    }
  }

  function scheduleReconnect() {
    clearReconnectTimer();
    // Backoff exponencial 1s, 2s, 4s, 8s... con tope en 30s.
    const delay = Math.min(BASE_BACKOFF_MS * 2 ** reconnectAttempts, MAX_BACKOFF_MS);
    reconnectAttempts += 1;
    reconnectTimer = setTimeout(() => get().connect(), delay);
  }

  function handleMessage(raw: string) {
    let msg: ServerMessage;
    try {
      msg = JSON.parse(raw) as ServerMessage;
    } catch {
      return; // mensaje no-JSON: ignorar
    }
    if (msg.type === "click_event") {
      set((s) => ({ lastEvents: [...s.lastEvents, msg.payload].slice(-MAX_EVENTS) }));
    } else if (msg.type === "error") {
      // Ej. suscripción a una URL ajena: lo mostramos como toast de error, no en silencio.
      useNotificationsStore
        .getState()
        .notify("error", msg.payload?.message ?? "Error de tiempo real.");
    }
  }

  function send(message: object) {
    const ws = get().socket;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(message));
    }
  }

  return {
    socket: null,
    isConnected: false,
    lastEvents: [],

    connect: () => {
      if (typeof window === "undefined") return;
      const token = useAuthStore.getState().accessToken;
      if (!token) return; // sin sesión no hay nada que conectar

      const existing = get().socket;
      if (
        existing &&
        (existing.readyState === WebSocket.OPEN ||
          existing.readyState === WebSocket.CONNECTING)
      ) {
        return; // ya conectado/conectando: idempotente
      }

      manualClose = false;
      const ws = new WebSocket(`${WS_URL}/ws?token=${encodeURIComponent(token)}`);

      ws.onopen = () => {
        reconnectAttempts = 0; // reconexión exitosa: resetea el backoff
        set({ isConnected: true });
      };
      ws.onmessage = (event) => handleMessage(event.data as string);
      ws.onclose = () => {
        set({ isConnected: false, socket: null });
        // Reconectar salvo que el cierre haya sido explícito (disconnect/logout).
        if (!manualClose) scheduleReconnect();
      };
      // onerror: el navegador siempre dispara onclose después; ahí va el reintento.
      ws.onerror = () => {};

      set({ socket: ws });
    },

    disconnect: () => {
      manualClose = true;
      clearReconnectTimer();
      reconnectAttempts = 0;
      const ws = get().socket;
      if (ws) {
        ws.onclose = null; // evita que el cierre dispare una reconexión
        ws.close();
      }
      set({ socket: null, isConnected: false });
    },

    subscribeToUrl: (shortCode) =>
      send({ type: "subscribe", payload: { shortCode } }),

    unsubscribeFromUrl: (shortCode) =>
      send({ type: "unsubscribe", payload: { shortCode } }),
  };
});

// Listener de sesión: conecta al iniciar sesión y desconecta al cerrarla, sin que
// auth importe este store (dependencia en una sola dirección). Solo en el cliente.
if (typeof window !== "undefined") {
  let prevToken = useAuthStore.getState().accessToken;
  useAuthStore.subscribe((state) => {
    const token = state.accessToken;
    if (!prevToken && token) useRealtimeStore.getState().connect();
    if (prevToken && !token) useRealtimeStore.getState().disconnect();
    prevToken = token;
  });
}
