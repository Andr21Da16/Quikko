export function Footer() {
  const year = new Date().getFullYear();
  return (
    <footer className="border-t border-zinc-200 bg-white py-10 dark:border-zinc-900 dark:bg-zinc-950">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-3 px-4 sm:flex-row sm:px-6">
        <span className="text-sm font-bold tracking-tight text-zinc-900 dark:text-zinc-50">
          Quikko<span className="text-accent-400">.</span>
        </span>
        <p className="text-sm text-zinc-500 dark:text-zinc-400">
          Acorta, comparte y mide tus enlaces en tiempo real.
        </p>
        <span className="text-sm text-zinc-400 dark:text-zinc-500">
          © {year} Quikko
        </span>
      </div>
    </footer>
  );
}
