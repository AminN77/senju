# IP and References — Using Competitors as UX References Responsibly

Senju is an open-source project. Its value grows as users migrate from commercial genomics platforms (DNAnexus is the primary UX reference for this phase). **"Make it familiar" does not mean "copy."** This document defines what is allowed, what is not, and what reviewers check for during the UI Foundation phase.

The spirit is simple: borrow **ideas**, build **from scratch**.

---

## 1. What is allowed (green)

- **Information architecture.** Putting projects, jobs, variants, admin, and settings at the same places in the nav where a DNAnexus user would expect them.
- **Mental model and workflow ordering.** Upload → QC → align → variant call → browse. This is the genomics workflow; it is not anyone's IP.
- **Domain terminology** drawn from the field (genomics, bioinformatics) and from widely adopted tool names (FASTQ, BWA, GATK, VCF, allele frequency, etc.).
- **Common UI patterns** that are not specific to any one vendor: dashboards with stat cards, timelines of pipeline stages, data tables with filter bars, log viewers with follow-tail. These are the grammar of operations UIs.
- **Studying competitor screenshots for layout ideas** — then closing the tab and building from scratch.
- **Public documentation** (open-access help articles, published tutorials) as **reference for domain semantics**, not as a source of copy.
- **Open-source reference projects** (Galaxy Project, open-source Nextflow plugins, etc.) under their declared licenses, with attribution per their license terms.

---

## 2. What is NOT allowed (red)

- **Copying CSS, component code, or class names** from any competitor's shipped site or app. Viewing DevTools is fine for study; pasting is not.
- **Copying or tracing icons, illustrations, or imagery** from any competitor. Our icon set is Lucide (MIT). Illustrations are original or sourced from projects with compatible licenses and attribution.
- **Copying microcopy, headings, error messages, tooltips, or onboarding text.** Paraphrasing a competitor's microcopy word-for-tweaked-word is still copying.
- **Reusing brand colors.** Our brand teal is original. If a color happens to be close to a competitor's, we adjust rather than defend the match.
- **Screenshots or exports from competitor products used as assets** (e.g. background imagery, example screens, anonymized dashboards).
- **Replicating trade dress** — the overall look-and-feel that makes a product identifiable. Side-by-side screenshots should make Senju's identity unmistakably distinct.
- **Reusing training data, model artifacts, or documentation** published under restrictive or proprietary terms.

---

## 3. The "third-party test"

Before merging any UI change, a reviewer should be able to answer **yes** to all of:

1. Could a neutral observer, viewing Senju's screen next to the competitor's, identify Senju as a distinct product?
2. Is every visual asset (icon, illustration, photo, font, sound) either original, or from an MIT / CC-BY / CC0 / OFL / Apache-2.0 source, with attribution where required?
3. Is every piece of copy on this screen written by the Senju team, not paraphrased from a competitor?
4. If this PR introduces a new dependency, is its license compatible with our distribution terms?

If any answer is **no**, the change does not merge until it is.

---

## 4. Acceptable reference materials

- **OpenAPI docs** from public standards bodies and published open-source references.
- **Academic and consortium resources** (ENSEMBL, UCSC, ClinVar, gnomAD documentation) for domain accuracy.
- **Open-source projects** (Galaxy, Nextflow, Cromwell, Snakemake) — their patterns and code are OSS-licensed and may be studied and (where appropriate and properly attributed) reused under their licenses.
- **Design inspiration from broadly practiced patterns** — Radix UI examples, shadcn/ui examples, open-source dashboards (Grafana, Superset, Metabase UIs are good pattern references and have permissive licenses on their code).

---

## 5. When in doubt

Escalate rather than ship. In PR review, if a reviewer cannot confidently answer the four questions in §3 for the change in front of them, they block the merge and tag a second reviewer. For structural questions (is this component architecture too close to a competitor's?), open an ADR — do not decide silently in a PR.

Suspected IP contamination that has already been merged is handled like a suspected secret leak: revert first, investigate after.

---

## 6. Attribution

Dependencies, fonts, icons, and any referenced open-source project are listed in `frontend/LICENSES.md` with their licenses. This file is generated or updated on every dependency change.

---

## 7. Automated guards

- `.cursor/rules/frontend-ip-guardrails.mdc` carries these rules forward into any AI-assisted work during the UI Foundation phase.
- PR template checklist includes an explicit "No assets or copy sourced from proprietary third parties" item; reviewers enforce.
- ESLint restricts arbitrary Tailwind color values (`text-[#...]`, `bg-[#...]`) which is a common vector for pasted styles.

---

## 8. What we do instead

For every screen, the design process is:

1. Look at 2–3 platforms (commercial and open-source) to understand the **problem** and the **mental model**.
2. Write the IA and flow in words first (`information-architecture.md`).
3. Wireframe the layout using our own primitives and tokens.
4. Build with our components — if the primitive doesn't exist, write it against our standards.
5. Review against the third-party test in §3.

This is slower than cloning. It is also how a project earns a promotable identity.
