# Accessibility

**WCAG 2.1 AA is the project minimum.** Every shipped route must meet it. This is a merge gate, not an aspiration.

Higher bar areas (admin, destructive actions, data entry) target **WCAG 2.2 AA** where the relevant success criterion applies.

---

## 1. Non-negotiables

1. **Keyboard-only operation.** Every interactive element reachable via `Tab` and operable via `Enter` / `Space` / arrow keys as appropriate. No mouse-only affordances.
2. **Visible focus.** Every focusable element has a visible focus ring (`border-focus` token) with 3:1 contrast against adjacent surfaces. `outline: none` is a lint error unless a compliant replacement is present in the same rule.
3. **Screen-reader labeling.** Every icon-only button has `aria-label`. Every form control has a programmatic label (visible `<label>` preferred; `aria-label` / `aria-labelledby` acceptable).
4. **Color is never the only signal.** Status, validation, and data-viz categories must also carry shape, text, or iconography.
5. **Reduced motion respected.** Every animation and transition collapses or degrades under `prefers-reduced-motion: reduce`.
6. **Contrast.** Enforced by token definitions and CI contrast test (see `color-palette.md` §7).

---

## 2. Semantics

- Use **HTML elements for their meaning**: `<button>` for buttons, `<a>` for navigation, `<nav>`, `<main>`, `<aside>`, `<header>`, `<footer>` for landmarks.
- One `<h1>` per page. Headings form a logical outline; never skip levels for styling.
- Lists use `<ul>` / `<ol>`. Tabular data uses `<table>`. Do not build tables out of divs.
- Form controls are wrapped in `<form>` and submit on `Enter`.

---

## 3. Focus management

- **Dialogs / modals** trap focus, return focus to trigger on close, and close on `Esc` (Radix handles this; do not override).
- **Route transitions** move focus to the page heading (`<h1 tabIndex={-1} ref={...}>`) and announce the new page via a live region.
- **Error states** move focus to the first invalid field on submit.
- **Skip link** ("Skip to main content") is the first focusable element on every page.

---

## 4. Live regions

- **Toasts** use `role="status"` (polite) for confirmations and `role="alert"` (assertive) for errors.
- **Async data loads** announce completion politely ("Job list updated").
- **Log viewer** auto-follow updates are announced every N seconds at most (never per line — that's a screen-reader flood).

---

## 5. Forms

- Labels are always present and visible. Placeholder text is not a label.
- Required fields are marked visually (`*`) **and** programmatically (`aria-required="true"`, `required` attribute).
- Errors are announced via `aria-describedby` on the input pointing to the error message; error messages use `role="alert"` on first appearance.
- Inline validation fires on **blur**, not per-keystroke (reduces screen-reader noise and flashing UI).

---

## 6. Data tables

- `<caption>` for table title (may be visually hidden but must exist).
- `<th scope="col">` and `<th scope="row">` where applicable.
- Sortable columns: `aria-sort` on the header, sort buttons are real `<button>`s.
- Row actions are buttons with `aria-label` including the row identity ("Delete job 4f8a…").
- Virtualized tables expose the total row count via `aria-rowcount` and each rendered row's position via `aria-rowindex`.

---

## 7. Data visualization

- Every chart has a **"View as table" toggle** rendering an accessible `<table>` with the same data.
- Chart titles and axis labels are present in both the visual and the screen-reader layer.
- Colors are not the only encoding — patterns, markers, or labels accompany color for series differentiation.
- `<svg>` gets `role="img"` and `<title>` + `<desc>` describing the chart.

---

## 8. Keyboard shortcuts

- Shortcut sequences never activate when focus is in an input (`<input>`, `<textarea>`, `[contenteditable]`).
- All shortcuts are discoverable via `?` cheat-sheet dialog.
- Every shortcut has an equivalent visible UI action.
- Shortcuts can be disabled in `/settings/preferences` (accommodates users of alternative input methods).

---

## 9. Testing

### Automated

- **axe-core** (`@axe-core/playwright`) runs against every route in CI on a built production bundle. Zero violations of `wcag21aa` tag.
- **jest-axe** runs against every component in its Storybook stories.
- **Contrast test** (`pnpm test:contrast`) asserts every token pair used in production CSS meets WCAG AA thresholds in both themes.
- **Lint** rules:
  - `jsx-a11y/alt-text`, `jsx-a11y/anchor-is-valid`, `jsx-a11y/aria-role`, `jsx-a11y/label-has-associated-control`, `jsx-a11y/no-autofocus`, `jsx-a11y/no-noninteractive-element-interactions`, all as **errors**.
  - Custom rule: no `outline: none` in CSS/Tailwind `outline-none` without a focus-visible replacement in the same block.

### Manual

Every feature PR must include a **keyboard-only walkthrough** in test evidence for the affected flow: tab order, focus visibility, `Enter`/`Space`/arrow keys, `Esc`.

Screen-reader spot checks (VoiceOver on macOS, NVDA on Windows) are required for dialogs, forms, and any new live region.

---

## 10. Known exemptions and how to request them

An a11y exemption requires:

1. A comment in the PR linking to the relevant WCAG success criterion.
2. A written rationale for why compliance is not feasible.
3. A dated TODO with an issue linked for remediation.
4. Reviewer sign-off explicitly on the exemption (not just the PR).

Exemptions are time-bounded (≤ 90 days). Unresolved exemptions past due are escalated in the next review.

---

## 11. Resources

- [WCAG 2.1 AA success criteria](https://www.w3.org/WAI/WCAG21/quickref/?versions=2.1&levels=aa)
- [Radix UI accessibility docs](https://www.radix-ui.com/primitives/docs/overview/accessibility)
- [Inclusive Components patterns](https://inclusive-components.design/)
- [axe-core rules reference](https://dequeuniversity.com/rules/axe/)
