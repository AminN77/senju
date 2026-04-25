import { ThemeSmokeToggle } from "@/components/theme-smoke-toggle";

const categories = [
  { label: "Color", sampleClass: "bg-brand-500 text-text-on-brand" },
  { label: "Spacing", sampleClass: "p-6" },
  { label: "Typography", sampleClass: "text-heading-md" },
  { label: "Radius", sampleClass: "rounded-lg" },
  { label: "Elevation", sampleClass: "shadow-2" },
  { label: "Motion", sampleClass: "transition-all duration-base ease-out-quad" },
  { label: "Z-index", sampleClass: "relative z-modal" },
] as const;

export default function TokensDemoPage() {
  return (
    <main className="mx-auto flex min-h-screen w-full max-w-5xl flex-col gap-6 px-6 py-12">
      <h1 className="text-heading-xl font-semibold text-text-primary">Design token demo</h1>
      <p className="text-body-md text-text-secondary">
        Representative token values across color, spacing, typography, radius, elevation, motion,
        and z-index.
      </p>

      <ThemeSmokeToggle />

      <section className="grid gap-4 md:grid-cols-2" aria-label="Token category samples">
        {categories.map((category) => (
          <article
            key={category.label}
            className={`rounded-md border border-border-default bg-surface-raised p-6 text-body-md text-text-primary ${category.sampleClass}`}
          >
            <h2 className="text-heading-sm font-medium text-text-primary">{category.label}</h2>
            <p className="mt-2 text-body-sm text-text-secondary">
              Sample utility usage: {category.sampleClass}
            </p>
          </article>
        ))}
      </section>
    </main>
  );
}
