import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// Protección de rutas (Agent 14). Intercepta la navegación ANTES de renderizar, así
// un usuario sin sesión no monta ni un frame de /dashboard/* (sin parpadeo ni llamadas
// de red del dashboard).
//
// NOTA DE SEGURIDAD: el middleware solo verifica que la cookie `quikko_session` EXISTA,
// no que el JWT sea válido o no haya expirado. La validación real ocurre en cada
// llamada del cliente HTTP (Agent 13), que maneja token expirado/inválido con su flujo
// de refresh o redirect. Esto es una barrera de UX; el backend sigue siendo la única
// fuente de verdad de autorización.
const SESSION_COOKIE = "quikko_session";

export function middleware(request: NextRequest) {
  const hasSession = Boolean(request.cookies.get(SESSION_COOKIE)?.value);
  const { pathname } = request.nextUrl;

  // Dashboard privado: sin cookie → /login.
  if (pathname.startsWith("/dashboard") && !hasSession) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // Redirect inverso: un usuario ya autenticado no debe ver login/register.
  if (hasSession && (pathname === "/login" || pathname === "/register")) {
    return NextResponse.redirect(new URL("/dashboard", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/dashboard/:path*", "/login", "/register"],
};
