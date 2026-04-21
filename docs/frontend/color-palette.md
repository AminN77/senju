# Color Palette

Senju's palette is original. It is **not** derived from any competitor's brand or UI colors. Values below define the token system; refinements to specific shades are permitted within an ADR or design review, but the structure and semantic naming are stable.

All values are expressed in the OKLCH color space at the source (via CSS `oklch()`), with sRGB hex fallbacks where useful for reference. OKLCH is the source of truth because perceptual uniformity matters for the data-viz scale and for guaranteeing dark/light parity.

---

## 1. Structure

The palette is organized into seven families:

1. **Neutral** — surfaces, borders, text (most UI pixels are neutral).
2. **Brand** — primary accent, used sparingly.
3. **Semantic** — success, warning, danger, info.
4. **Data viz** — categorical (10), sequential, diverging.
5. **Surface tokens** — theme-aware surface roles.
6. **Text tokens** — theme-aware text roles.
7. **Border tokens** — theme-aware border roles.

Neutrals and brand each have a 50→900 ramp (50 lightest, 900 darkest). Semantic colors have `subtle`, `muted`, `solid`, `solid-hover`, `on-solid`, `border`.

---

## 2. Brand

Brand is a deep, slightly cool teal — distinctive in genomics/ops tooling, legible on both dark and light surfaces, and intentionally not matching any major commercial competitor's hue.

| Token | OKLCH | sRGB (reference) | Role |
| --- | --- | --- | --- |
| `brand-50` | `oklch(96% 0.02 200)` | `#ecf7f6` | Brand-tinted surface |
| `brand-100` | `oklch(92% 0.04 200)` | `#d2ebe9` | Hovered brand-tinted surface |
| `brand-200` | `oklch(86% 0.07 200)` | `#a9dad6` | Subtle borders |
| `brand-300` | `oklch(78% 0.10 200)` | `#78c3bd` | Disabled brand |
| `brand-400` | `oklch(68% 0.13 200)` | `#40a59d` | Secondary accents |
| `brand-500` | `oklch(60% 0.14 200)` | `#1f8680` | **Primary brand** |
| `brand-600` | `oklch(52% 0.13 200)` | `#116b66` | Hovered primary |
| `brand-700` | `oklch(44% 0.11 200)` | `#0b544f` | Pressed primary |
| `brand-800` | `oklch(36% 0.09 200)` | `#083f3b` | Deep brand surface |
| `brand-900` | `oklch(28% 0.07 200)` | `#052c29` | Deepest brand accent |

Usage rules:

- Primary CTAs use `brand-500` fill + `on-solid` text.
- Large brand surfaces use `brand-50` / `brand-900` (theme-aware via surface tokens).
- Never apply `brand-500` as body text; contrast is borderline at small sizes.

---

## 3. Neutrals

Warm-neutral ramp with a very slight blue cast (cooler than gray, warmer than slate). Chosen to feel calm in long data-viewing sessions.

| Token | OKLCH | sRGB (reference) |
| --- | --- | --- |
| `neutral-50` | `oklch(98% 0.003 250)` | `#f8f9fb` |
| `neutral-100` | `oklch(96% 0.004 250)` | `#f1f3f6` |
| `neutral-200` | `oklch(92% 0.006 250)` | `#e4e7ec` |
| `neutral-300` | `oklch(85% 0.008 250)` | `#cdd2da` |
| `neutral-400` | `oklch(72% 0.010 250)` | `#a0a7b3` |
| `neutral-500` | `oklch(58% 0.012 250)` | `#74798a` |
| `neutral-600` | `oklch(48% 0.013 250)` | `#595e6d` |
| `neutral-700` | `oklch(38% 0.014 250)` | `#404654` |
| `neutral-800` | `oklch(28% 0.014 250)` | `#282e3a` |
| `neutral-900` | `oklch(20% 0.014 250)` | `#171c27` |
| `neutral-950` | `oklch(14% 0.013 250)` | `#0d1119` |

---

## 4. Semantic colors

Each semantic color has six roles. Dark/light values are derived from the same OKLCH seed with different lightness.

### Success (green)

| Role | Dark theme | Light theme |
| --- | --- | --- |
| `success-subtle` | `oklch(22% 0.04 150)` | `oklch(96% 0.03 150)` |
| `success-muted` | `oklch(30% 0.06 150)` | `oklch(90% 0.06 150)` |
| `success-solid` | `oklch(65% 0.15 150)` | `oklch(50% 0.15 150)` |
| `success-solid-hover` | `oklch(70% 0.15 150)` | `oklch(45% 0.15 150)` |
| `success-on-solid` | `oklch(98% 0.01 150)` | `oklch(98% 0.01 150)` |
| `success-border` | `oklch(40% 0.08 150)` | `oklch(82% 0.09 150)` |

### Warning (amber)

Seed hue: 70. Same role structure.

### Danger (red)

Seed hue: 25. Same role structure.

### Info (blue, distinct from brand teal)

Seed hue: 250. Same role structure.

Full per-role OKLCH tables for warning/danger/info are generated from the same formulas in `frontend/src/styles/tokens.css`. The source of truth is code; this doc captures the rules.

---

## 5. Data-viz palette

### Categorical (10 colors)

Designed to be perceptually distinguishable, robust to common color-vision deficiencies (verified against deuteranopia and protanopia simulators), and ordered so the first 3–5 colors alone still separate clearly.

| Token | Role | Approx hue |
| --- | --- | --- |
| `dv-1` | Primary series | 200 (teal, brand-adjacent) |
| `dv-2` | Secondary series | 30 (warm orange) |
| `dv-3` | Tertiary series | 280 (purple) |
| `dv-4` | | 140 (green) |
| `dv-5` | | 10 (red) |
| `dv-6` | | 60 (gold) |
| `dv-7` | | 320 (magenta) |
| `dv-8` | | 180 (cyan) |
| `dv-9` | | 100 (olive) |
| `dv-10` | | 240 (indigo) |

Lightness is tuned per theme so that all 10 pass 3:1 contrast against the background in both dark and light modes.

### Sequential

Single-hue ramp keyed off brand teal for density/intensity encoding (`dv-seq-1` through `dv-seq-9`).

### Diverging

Two-hue diverging scale centered on neutral — blue (cold) → white (zero) → red (hot). For signed values like log-fold change or z-scores.

---

## 6. Surface / text / border roles (theme-aware)

These are the tokens component code actually uses. They remap per theme.

### Surface tokens

| Token | Dark theme source | Light theme source | Use |
| --- | --- | --- | --- |
| `surface-base` | `neutral-950` | `neutral-50` | Page background |
| `surface-raised` | `neutral-900` | `#ffffff` | Cards, panels |
| `surface-overlay` | `neutral-800` | `#ffffff` | Dialogs, popovers |
| `surface-sunken` | `neutral-950` | `neutral-100` | Inset areas (code blocks, inputs) |
| `surface-inverse` | `neutral-50` | `neutral-900` | Inverted callouts |
| `surface-brand-subtle` | `brand-900` at ~40% mix with `neutral-900` | `brand-50` | Brand-tinted zones |

### Text tokens

| Token | Dark | Light | Use |
| --- | --- | --- | --- |
| `text-primary` | `neutral-50` | `neutral-900` | Body, headings |
| `text-secondary` | `neutral-300` | `neutral-600` | Secondary body |
| `text-muted` | `neutral-400` | `neutral-500` | Metadata, captions |
| `text-disabled` | `neutral-600` | `neutral-400` | Disabled state |
| `text-on-brand` | `neutral-50` | `neutral-50` | Text on brand-solid fills |
| `text-on-inverse` | `neutral-900` | `neutral-50` | Text on inverted surface |
| `text-link` | `brand-400` | `brand-600` | Inline links |
| `text-danger` | `danger-solid` | `danger-solid` | Inline destructive text |

### Border tokens

| Token | Dark | Light | Use |
| --- | --- | --- | --- |
| `border-subtle` | `neutral-800` | `neutral-200` | Dividers within a surface |
| `border-default` | `neutral-700` | `neutral-300` | Card / input borders |
| `border-strong` | `neutral-600` | `neutral-400` | Hovered inputs |
| `border-focus` | `brand-500` | `brand-500` | Focus rings (paired with offset) |

---

## 7. Contrast compliance (WCAG 2.1 AA)

Every token pair used for text-on-surface or interactive element must meet:

- **4.5:1** for body text (< 18px regular, < 14px bold).
- **3:1** for large text (≥ 18px regular, ≥ 14px bold) and non-text UI (borders, focus rings, icons conveying state).
- **3:1** between adjacent data-viz series against the chart background.

Representative pairings (both themes must pass):

| Foreground | Background | Target | Status |
| --- | --- | --- | --- |
| `text-primary` | `surface-base` | 4.5:1 | ✔ (verified in CI) |
| `text-secondary` | `surface-base` | 4.5:1 | ✔ |
| `text-muted` | `surface-base` | 4.5:1 for body, 3:1 for captions ≥ 12px | ✔ for captions only — do not use for body |
| `text-on-brand` | `brand-500` | 4.5:1 | ✔ |
| `text-primary` | `surface-raised` | 4.5:1 | ✔ |
| `border-focus` | `surface-base` | 3:1 | ✔ |
| `success-solid` on `surface-base` | — | 3:1 (non-text) | ✔ |

A contrast-verification test (`pnpm test:contrast`) runs every token pair used in production CSS against WCAG thresholds in CI. A failure blocks merge.

---

## 8. Status → color mapping (job and variant UI)

| State | Token family | Use |
| --- | --- | --- |
| `queued` | `neutral-400` / `neutral-500` | Awaiting start |
| `running` | `info-solid` + pulse animation | In progress |
| `succeeded` | `success-solid` | Terminal success |
| `failed` | `danger-solid` | Terminal failure |
| `canceled` | `neutral-500` | User-canceled |
| `paused` / `checkpointed` | `warning-solid` | Held (aligned with backend reliability ADR) |

---

## 9. What the palette is NOT

- It is not DNAnexus's palette, nor Seven Bridges', nor Terra's, nor Seqera's.
- It does not reuse any brand hex value from a competitor's published brand guidelines.
- The brand teal is distinct enough from any major competitor's accent that side-by-side screenshots make identification trivial.

If a future design revision wants to adjust brand hue, that decision requires an ADR update, not a one-off token edit.
