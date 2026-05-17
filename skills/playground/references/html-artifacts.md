# HTML Artifact Guidance

## Choose the artifact shape

Use the smallest shape that creates the intended behavior:

| Request | Shape |
|---|---|
| Compare N approaches / designs | Side-by-side comparison columns + Submit |
| Explore design directions | Live-rendered options user can point at |
| Plan / spec / RFC | Milestone timeline + data flow + risk table |
| Code review / PR | Annotated diff with margin notes + severity |
| Module map | Boxes-and-arrows SVG with hot-path highlight |
| Design tokens | Swatches + type specimens + copy-on-click |
| Component variants | Contact sheet: all sizes/states/intents |
| Parameter tuning | Sliders/knobs + live preview + Submit |
| Clickable prototype | Linked screens, real interaction |
| Slide deck | `<section>` slides + keyboard nav + View Transitions |
| Feature / concept explainer | TL;DR box + collapsible sections + tabbed code |
| Status report | Chart + colored timeline + what-shipped list |
| Incident timeline | Minute-by-minute log + follow-up checklist |
| Ticket triage | Drag-and-drop kanban + Submit ordering |
| Config / flag editor | Grouped fields + dependency warnings + diff export |
| Prompt tuner | Editable template left + live preview right |
| Workshop lab | Step cards + copy prompts + paste-back + export |
| Manager summary | Recommendation first + 3 cards + one CTA |
| Data explorer | Filterable table + Popover row detail + CSV export |
| Mind map | Branching SVG tree + collapse/expand + Submit |
| Timeline / roadmap | Gantt bars + milestone Popover + export |

## Reference examples

All 20 are in `$SKILL_DIR/examples/`. Read the matching file before producing that artifact type.

**Exploration & Planning**
- `01-exploration-code-approaches.html` — three approaches side-by-side, tradeoffs inline
- `02-exploration-visual-designs.html` — rendered layout/palette options
- `16-implementation-plan.html` — milestones, data flow, mockups, risk table

**Code Review**
- `03-code-review-pr.html` — annotated diff with margin notes, severity tags, jump links
- `17-pr-writeup.html` — author's view: motivation, file-by-file tour, where to focus
- `04-code-understanding.html` — package as boxes-and-arrows, hot path highlighted

**Design**
- `05-design-system.html` — tokens rendered as swatches, copy-on-click
- `06-component-variants.html` — all sizes, states, and intents on one sheet

**Prototyping**
- `07-prototype-animation.html` — animation in isolation with duration/easing sliders
- `08-prototype-interaction.html` — four linked screens, real click-through

**Diagrams**
- `10-svg-illustrations.html` — SVG figures for a post, tweak and copy individually
- `13-flowchart-diagram.html` — deploy pipeline, click step for details

**Decks**
- `09-slide-deck.html` — arrow-key navigation, no build step

**Research**
- `14-research-feature-explainer.html` — TL;DR, collapsible steps, tabbed snippets, FAQ
- `15-research-concept-explainer.html` — live interactive ring, comparison table, glossary

**Reports**
- `11-status-report.html` — what shipped/slipped, small chart, Monday-morning skim
- `12-incident-report.html` — minute-by-minute timeline, log excerpts, follow-up checklist

**Custom Editors**
- `18-editor-triage-board.html` — drag tickets to Now/Next/Later/Cut, copy ordering
- `19-editor-feature-flags.html` — toggles with dependency warnings, copy diff
- `20-editor-prompt-tuner.html` — editable template left, three sample inputs right

## Design system preference

For one-off documents and manager-facing artifacts, use the "HTML effectiveness" visual language. Implemented with Tailwind custom tokens:

```html
<style type="text/tailwindcss">
  @theme {
    --color-ivory: #FAF9F5;
    --color-slate-dark: #141413;
    --color-clay: #D97757;
    --color-oat: #E3DACC;
    --color-olive: #788C5D;
    --color-gray-150: #F0EEE6;
    --color-gray-300: #D1CFC5;
    --color-gray-500: #87867F;
    --color-gray-700: #3D3D3A;
  }
</style>
```

Type: Tailwind's `font-serif` for main headlines and card titles; `font-sans` for body; `font-mono` for labels, status, and actions.

## Scannability rules

- Put recommendation before options when the recipient wants a decision.
- Max 3 primary choices per view.
- Cards: title, one-line value prop, 3 bullets, one action.
- Use Popover or `<details>` for progressive disclosure.
- One primary CTA at end.
- If asked to "strip it down" — cut background context before cutting decisions.

## Per-type guidance

### Slide decks
Keyboard nav: `→`/`Space` next, `←` previous, `F` fullscreen, `Esc` exit. Hash-based deep links (`#3`). Progress counter in corner. Fixed 16:9 ratio, letterboxed. Slide types: title, one-idea, code, diagram, comparison (use sparingly), demo (live HTML in slide), section break, recap. Speaker notes toggled with `N` — written in `<aside>` inside each `<section>`. Use View Transitions on slide change. Pick one aesthetic (editorial / engineering / brutalist / documentary) and commit.

### Throwaway editors
Pre-populate with the user's actual data — never make them paste it again. Export is non-negotiable; always end with a Submit or Copy button. Patterns: drag-and-drop board (kanban triage), form-based config with dependency warnings, side-by-side prompt/template editor, dataset curator (approve/reject with keyboard shortcuts `j`/`k`/`y`/`n`), annotation tool. Keyboard shortcuts for anything involving more than ~10 items. Show counts ("37 to review, 12 approved"). For drag-and-drop reordering, use `x-sort` / `@alpinejs/sort` — see `alpine-plugins.md` § Sort.

### Interactive playgrounds
Real-time updating preview as the user manipulates controls. Sliders/dropdowns/toggles; always show current value as text next to the control with units ("220ms", "1.04×"). Submit sends tuned values back. Patterns: single-component tuner (preview + controls), algorithm parameter explorer (synthetic visualization of behavior), multi-parameter sweep (grid of combinations).

### Comparison / brainstorm grid
N variants side-by-side (2–4 columns). Each column: variant title, the rendered thing, pro/con/tradeoff callouts. Submit sends the selected variant + rationale. Don't use this for things that require sequential reading — use a spec doc instead.

### Spec / planning docs
Milestones on a visual timeline with package/surface tags. Data flow diagram (inline SVG). Inline mockups. Risk table. Open questions section. Timestamp in footer. Skimmable on a phone in a meeting.

### Code review artifacts
Annotated diff with margin notes, severity tags, jump links. PR writeup: motivation, before/after, file-by-file tour with the *why*, where to focus review. Module map: boxes-and-arrows with hot path highlighted, entry points listed.

### Diagrams / flowcharts
Inline SVG — real `<g>` and `<path>`, not PNGs. SVG text in `<foreignObject>` for variable-length labels. Annotated flowchart: click any node to see details in a Popover. Architecture map: zoom/pan with CSS `transform`. Sequence diagrams: time flows top-to-bottom, actors as columns.

### Research explainers / reports
TL;DR box at top. Collapsible sections: prefer `x-collapse` from `@alpinejs/collapse` over `x-show` + manual height CSS. Fall back to `<details>` for low-stakes disclosure. Tabbed code samples. Glossary with hover definitions (Popover). Status reports: chart for key metric (inline SVG or CSS bar), colored timeline. Incident reports: minute-by-minute timeline, log excerpts, follow-up checklist.

### Data explorer
Filterable table with faceted search. `x-model` bound search input + computed filter. Column sort. Row click opens detail in Popover or `<dialog>`. Export filtered view as CSV or JSON. Log viewers: monospace, line numbers, severity coloring.

### Design tokens showcase
Color swatches with hex/var displayed. Type scale specimens. Spacing scale with visual bars. Copy-on-click for any token value.

### Mind map
Branching SVG tree, centered root. Click node to expand/collapse. Submit sends the full tree as a nested JSON structure.

### Timeline / roadmap
Gantt-style: swim lanes per team/area, bars proportional to duration, dependency arrows. Click milestone to open detail in Popover. Export as markdown table.

### Local-first app
Use `templates/local-first-app/` as the starting point — copy it, don't reconstruct from scratch. Architecture: SQLite via PGlite + OPFS. Worker thread handles all DB ops via `{ action, payload }` postMessage; Alpine wires UI to worker. Features already scaffolded: timer, manual entry, period filter, project summary, entries list, edit dialog, CSV/JSON export, toast notifications.

## GitHub Pages delivery

For a shareable static lab:
- `index.html` in repo root (or `/docs` folder)
- Avoid private/internal data in public repos
- Include README with "open locally" and deploy steps
- All state local unless collaboration is explicitly required
- Tailwind CDN + Alpine CDN = no build step needed
