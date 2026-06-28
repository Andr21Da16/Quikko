import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

// cn compone clases condicionalmente (clsx) y resuelve conflictos de Tailwind
// (tailwind-merge): así un caller puede pasar className y sobreescribir estilos sin
// que queden dos utilidades en pugna (ej. `p-2` + `p-4`).
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
