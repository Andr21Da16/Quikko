// Puente de sesión entre el cliente HTTP y el store de auth.
//
// El cliente necesita leer/actualizar los tokens, pero NO debe importar el store
// directamente (crearía un ciclo: client → store → endpoints → client) ni tocar
// localStorage. En su lugar, el store de auth se "registra" aquí al inicializarse y
// el cliente consume estos accessors. session.ts no importa nada → rompe el ciclo.

export type SessionAccessors = {
  getAccessToken: () => string | null;
  getRefreshToken: () => string | null;
  // setTokens lo usa el silent refresh para guardar el nuevo access token.
  setTokens: (accessToken: string, refreshToken: string) => void;
  // clear lo usa el cliente cuando el refresh falla definitivamente (re-login).
  clear: () => void;
};

let accessors: SessionAccessors | null = null;

export function registerSession(a: SessionAccessors): void {
  accessors = a;
}

export const session = {
  getAccessToken: (): string | null => accessors?.getAccessToken() ?? null,
  getRefreshToken: (): string | null => accessors?.getRefreshToken() ?? null,
  setTokens: (accessToken: string, refreshToken: string): void =>
    accessors?.setTokens(accessToken, refreshToken),
  clear: (): void => accessors?.clear(),
};
