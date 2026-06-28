# Protocolo WebSocket en tiempo real (`GET /ws`)

OpenAPI 3.0 no modela WebSockets, así que el endpoint `GET /ws` se documenta aquí
en vez de en `openapi.yaml`. Refleja la implementación real de `internal/realtime/`
(Agent 6). Para el diseño y las decisiones, ver `docs/specs/06-agent-realtime.md`.

## Conexión

```
ws://localhost:8080/ws?token=<accessToken>
```

- El **JWT de acceso** viaja como query param `token` (no por header `Authorization`,
  porque la API de WebSocket del navegador no permite headers personalizados en el
  handshake).
- El token se valida **antes** del upgrade. Si falta o es inválido/expirado, el server
  responde un `401` HTTP normal (envelope JSON) y **no** hace el upgrade:
  - `AUTH_TOKEN_INVALID` — token ausente o inválido.
  - `AUTH_TOKEN_EXPIRED` — token expirado.
- Al conectar, el cliente queda **suscrito automáticamente** al room de su usuario
  (`user:{userID}`), por lo que recibe los clics de todas sus URLs sin pedir nada.

## Formato de mensaje (sobre común)

Todo mensaje, en ambos sentidos, es un JSON con esta forma:

```json
{ "type": "<tipo>", "payload": { ... } }
```

`type` es uno de: `subscribe`, `unsubscribe`, `click_event`, `error`.

## Mensajes del cliente al servidor

### `subscribe`
Suscribe al cliente al canal de una URL concreta (`url:{shortCode}`), previa
validación de ownership. Útil para una vista de detalle de una URL.

```json
{ "type": "subscribe", "payload": { "shortCode": "xYz12A" } }
```

Si la URL no es del usuario (o no existe), el server responde un mensaje `error`
(no cierra la conexión).

### `unsubscribe`
Cancela la suscripción a una URL concreta.

```json
{ "type": "unsubscribe", "payload": { "shortCode": "xYz12A" } }
```

## Mensajes del servidor al cliente

### `click_event`
Se emite por cada clic en una URL a la que el cliente está suscrito (su room de
usuario o un room de URL).

```json
{
  "type": "click_event",
  "payload": {
    "shortCode": "xYz12A",
    "country": "PE",
    "deviceType": "desktop",
    "browser": "Chrome",
    "timestamp": "2026-06-27T12:34:56Z"
  }
}
```

### `error`
Comunica un error de protocolo (ej. suscripción a una URL ajena) sin cerrar la
conexión.

```json
{ "type": "error", "payload": { "code": "FORBIDDEN", "message": "..." } }
```

## Notas de implementación

- El broadcast del hub es **no bloqueante**: si el buffer de un cliente está lleno,
  ese mensaje se descarta para no bloquear al resto.
- El canal `send` del cliente nunca se cierra desde fuera (se usa un `done` +
  `sync.Once`), así un broadcast concurrente jamás escribe en un canal cerrado.
