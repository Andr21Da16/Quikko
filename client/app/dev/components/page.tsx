import { notFound } from "next/navigation";
import { ComponentsShowcase } from "./showcase";

// Página interna de componentes
export default function DevComponentsPage() {
  if (process.env.NODE_ENV === "production") {
    notFound();
  }
  return <ComponentsShowcase />;
}
