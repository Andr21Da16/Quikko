"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { Menu, X } from "lucide-react";
import { cn } from "@/lib/utils/cn";
import { CtaLink } from "./CtaLink";

// Nav de la landing. Sticky, translúcida con blur al hacer scroll.
// Deliberadamente austera (pocos elementos), como las referencias. En mobile los links
// centrales se colapsan en un menú hamburguesa simple; el CTA primario queda visible.
const LINKS = [
  { href: "#features", label: "Características" },
  { href: "#how-it-works", label: "Cómo funciona" },
  { href: "#pricing", label: "Precios" },
];

export function LandingNav() {
  const [scrolled, setScrolled] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <header
      className={cn(
        "sticky top-0 z-50 transition-colors",
        scrolled
          ? "border-b border-zinc-200 bg-white/70 backdrop-blur-md dark:border-zinc-800 dark:bg-zinc-950/70"
          : "border-b border-transparent",
      )}
    >
      <nav className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
        <Link
          href="/"
          className="text-lg font-bold tracking-tight text-zinc-900 dark:text-zinc-50"
        >
          Quikko<span className="text-accent-400">.</span>
        </Link>

        {/* Links centrales (desktop) */}
        <div className="hidden items-center gap-8 md:flex">
          {LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="text-sm text-zinc-600 transition-colors hover:text-zinc-900 dark:text-zinc-300 dark:hover:text-white"
            >
              {link.label}
            </a>
          ))}
        </div>

        {/* CTAs (desktop) */}
        <div className="hidden items-center gap-2 md:flex">
          <CtaLink href="/login" variant="ghost">
            Iniciar sesión
          </CtaLink>
          <CtaLink href="/register" variant="primary">
            Crear cuenta gratis
          </CtaLink>
        </div>

        {/* Mobile: CTA primario + hamburguesa */}
        <div className="flex items-center gap-2 md:hidden">
          <CtaLink href="/register" variant="primary" className="h-9 px-3">
            Crear cuenta
          </CtaLink>
          <button
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            aria-label={menuOpen ? "Cerrar menú" : "Abrir menú"}
            aria-expanded={menuOpen}
            className="inline-flex size-9 items-center justify-center rounded-lg text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800"
          >
            {menuOpen ? <X className="size-5" /> : <Menu className="size-5" />}
          </button>
        </div>
      </nav>

      {/* Menú mobile desplegable */}
      {menuOpen && (
        <div className="border-t border-zinc-200 bg-white px-4 py-3 md:hidden dark:border-zinc-800 dark:bg-zinc-950">
          <div className="flex flex-col gap-1">
            {LINKS.map((link) => (
              <a
                key={link.href}
                href={link.href}
                onClick={() => setMenuOpen(false)}
                className="rounded-lg px-3 py-2 text-sm text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800"
              >
                {link.label}
              </a>
            ))}
            <a
              href="/login"
              onClick={() => setMenuOpen(false)}
              className="rounded-lg px-3 py-2 text-sm font-medium text-zinc-700 hover:bg-zinc-100 dark:text-zinc-200 dark:hover:bg-zinc-800"
            >
              Iniciar sesión
            </a>
          </div>
        </div>
      )}
    </header>
  );
}
