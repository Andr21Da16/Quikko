"use client";

import { motion } from "framer-motion";
import type { ReactNode } from "react";

// Entrada sutil al hacer scroll: fade + ligero slide-up, una sola vez.
// Timing rápido y discreto (~0.45s, easeOut) como en las referencias (Linear/Vercel),
// no un efecto largo y llamativo. `delay` permite escalonar items de un grid.
export function Reveal({
  children,
  delay = 0,
  className,
}: {
  children: ReactNode;
  delay?: number;
  className?: string;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, margin: "-80px" }}
      transition={{ duration: 0.45, ease: "easeOut", delay }}
      className={className}
    >
      {children}
    </motion.div>
  );
}
