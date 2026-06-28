"use client";

import { type FormEvent, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Lock, Mail } from "lucide-react";
import { Button, Input, Spinner } from "@/components/ui";
import { AuthShell } from "@/components/auth/AuthShell";
import { useAuthStore } from "@/store/auth";
import { ApiError } from "@/lib/api/client";

// Página de registro real. Hermana directa de /login: reutiliza el
// mismo AuthShell, misma estructura y mismo patrón de hidratación/errores para que la
// transición entre ambas se sienta consistente.
//
// El acceso ya está filtrado por middleware.ts: un usuario con cookie de sesión es
// redirigido a /dashboard antes de renderizar. El campo de confirmación de password es
// SOLO de cliente (el backend no lo pide): no se envía, solo evita errores de tipeo.

// Validación de email básica en cliente (la verdad la valida el backend).
const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
// El backend exige password de al menos 8 caracteres (auth.Register).
const MIN_PASSWORD = 8;

export default function RegisterPage() {
  const router = useRouter();
  const register = useAuthStore((s) => s.register);
  const isLoading = useAuthStore((s) => s.isLoading);

  const emailRef = useRef<HTMLInputElement>(null);

  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [emailError, setEmailError] = useState<string>();
  const [passwordError, setPasswordError] = useState<string>();
  const [confirmError, setConfirmError] = useState<string>();
  const [formError, setFormError] = useState<string>();

  // Anti-parpadeo: arranca en `false` (el server no accede a localStorage/persist) y se
  // resuelve en el efecto, ya en cliente. Mismo enfoque que /login.
  const [hydrated, setHydrated] = useState(false);

  useEffect(() => {
    // setState siempre en un callback (no síncrono en el cuerpo del efecto), como exige
    // react-hooks/set-state-in-effect.
    const markHydrated = () => setHydrated(true);
    const unsub = useAuthStore.persist.onFinishHydration(markHydrated);
    if (useAuthStore.persist.hasHydrated()) queueMicrotask(markHydrated);
    return unsub;
  }, []);

  // Caso límite que el middleware no atajó: si ya hay sesión tras rehidratar, al dashboard.
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

    // Validaciones de cliente: bloquean el submit antes de llamar a la API.
    const emailInvalid = !EMAIL_RE.test(email);
    const passwordInvalid = password.length < MIN_PASSWORD;
    const confirmInvalid = confirm !== password;

    setEmailError(emailInvalid ? "Introduce un email válido." : undefined);
    setPasswordError(
      passwordInvalid
        ? `La contraseña debe tener al menos ${MIN_PASSWORD} caracteres.`
        : undefined,
    );
    setConfirmError(confirmInvalid ? "Las contraseñas no coinciden." : undefined);

    if (emailInvalid || passwordInvalid || confirmInvalid) {
      if (emailInvalid) emailRef.current?.focus();
      return;
    }

    try {
      // register en el store hace login implícito (guarda tokens + cookie de sesión).
      await register(email, password);
      router.push("/dashboard");
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === "AUTH_EMAIL_TAKEN") {
          // Error específico cerca del campo de email, no genérico.
          setEmailError("Ya existe una cuenta con este email.");
          emailRef.current?.focus();
        } else {
          setFormError(err.message);
        }
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
      title="Crear cuenta"
      subtitle="Empieza a acortar y medir tus enlaces"
      footer={
        <>
          ¿Ya tienes cuenta?{" "}
          <Link
            href="/login"
            className="font-medium text-brand-600 hover:text-brand-700 dark:text-brand-400 dark:hover:text-brand-300"
          >
            Iniciar sesión
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
          autoComplete="new-password"
          placeholder="Mínimo 8 caracteres"
          icon={<Lock size={16} />}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          error={passwordError}
          required
        />

        <Input
          label="Confirmar contraseña"
          type="password"
          name="confirmPassword"
          autoComplete="new-password"
          placeholder="••••••••"
          icon={<Lock size={16} />}
          value={confirm}
          onChange={(e) => setConfirm(e.target.value)}
          error={confirmError}
          required
        />

        <Button type="submit" isLoading={isLoading} className="mt-1 w-full">
          {isLoading ? "Creando cuenta…" : "Crear cuenta"}
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
