"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Lock, Mail } from "lucide-react";
import { Button, Input, Spinner } from "@/components/ui";
import { AuthShell } from "@/components/auth/AuthShell";
import { useAuthStore } from "@/store/auth";
import { ApiError } from "@/lib/api/client";

// Página de login real. Consume el store de auth  end-to-end contra
// el backend. El acceso a esta página ya está filtrado por middleware.ts: un usuario con
// cookie de sesión es redirigido a /dashboard ANTES de renderizar, así que aquí solo se
// llega sin sesión activa.
//

// Validación de email básica en cliente (el backend valida la verdad).
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export default function LoginPage() {
  const router = useRouter();
  const login = useAuthStore((s) => s.login);
  const isLoading = useAuthStore((s) => s.isLoading);

  const emailRef = useRef<HTMLInputElement>(null);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [emailError, setEmailError] = useState<string>();
  const [formError, setFormError] = useState<string>();

  // Anti-parpadeo: no mostramos el formulario hasta saber el estado de sesión rehidratado.
  // Arranca en `false` (el server no tiene acceso a localStorage/persist) y se resuelve
  // en el efecto, ya en cliente.
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    // setState va siempre en un callback (no síncrono en el cuerpo del efecto), tal como
    // exige react-hooks/set-state-in-effect.
    const markHydrated = () => setHydrated(true);
    const unsub = useAuthStore.persist.onFinishHydration(markHydrated);
    if (useAuthStore.persist.hasHydrated()) queueMicrotask(markHydrated);
    return unsub;
  }, []);

  // Si tras rehidratar resulta que ya hay sesión (caso límite que el middleware no atajó),
  // no mostramos el login: vamos al dashboard.
  useEffect(() => {
    if (hydrated && useAuthStore.getState().isAuthenticated) {
      router.replace("/dashboard");
    }
  }, [hydrated, router]);

  // Foco inicial en el campo de email (accesibilidad).
  useEffect(() => {
    if (hydrated) emailRef.current?.focus();
  }, [hydrated]);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setFormError(undefined);
    setEmailError(undefined);

    if (!EMAIL_RE.test(email)) {
      setEmailError("Introduce un email válido.");
      emailRef.current?.focus();
      return;
    }

    try {
      await login(email, password);
      router.push("/dashboard");
    } catch (err) {
      // Mensaje del backend mapeado a algo legible. Se muestra inline (no por toast) para
      // no duplicar el mismo error en dos sitios de la pantalla.
      if (err instanceof ApiError) {
        setFormError(
          err.code === "AUTH_INVALID_CREDENTIALS"
            ? "Email o contraseña incorrectos."
            : err.message,
        );
      } else {
        setFormError("Ocurrió un error inesperado. Inténtalo de nuevo.");
      }
    }
  }

  if (!hydrated) {
    return (
      <main className="flex min-h-dvh items-center justify-center bg-white dark:bg-zinc-950">
        <Spinner size={28} className="text-brand-600 dark:text-brand-400" />
      </main>
    );
  }

  return (
    <AuthShell
      title="Iniciar sesión"
      subtitle="Bienvenido de vuelta a Quikko"
      footer={
        <>
          ¿No tienes cuenta?{" "}
          <Link
            href="/register"
            className="font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
          >
            Crear una
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
        <Input
          ref={emailRef}
          label="Email"
          type="email"
          name="email"
          autoComplete="email"
          placeholder="tu@email.com"
          icon={<Mail size={16} />}
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          error={emailError}
          required
        />

        <Input
          label="Contraseña"
          type="password"
          name="password"
          autoComplete="current-password"
          placeholder="••••••••"
          icon={<Lock size={16} />}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          required
        />

        <Button type="submit" isLoading={isLoading} className="mt-1 w-full">
          {isLoading ? "Entrando…" : "Iniciar sesión"}
        </Button>

        {formError && (
          <p
            role="alert"
            className="text-center text-sm text-red-600 dark:text-red-400"
          >
            {formError}
          </p>
        )}
      </form>
    </AuthShell>
  );
}
