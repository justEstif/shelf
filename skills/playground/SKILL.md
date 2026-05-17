---
name: playground
description: "Create polished local-first HTML playgrounds and micro-apps for exploration, workshops, demos, and manager-ready artifacts. Use when the user asks for a playground, interactive HTML artifact, static GitHub Pages lab, self-guided exercise, visual explorer, copy/paste AI workflow, slide deck, diagram, data explorer, code review artifact, comparison matrix, timeline, throwaway editor, brainstorm grid, mind map, design token showcase, spec/planning doc, or HTML instead of markdown/slides. Keywords: playground, micro-app, static HTML, GitHub Pages, Alpine, Tailwind CDN, Popover API, View Transitions, local-first, copy prompt, export markdown, slide deck, kanban, ERD, mind map, data explorer, spec, implementation plan."
---

# Playground

Build self-contained browser artifacts that help users *try the idea*, not read about it. Optimize for a URL-sharable, local-first experience: static files, browser state, copy/paste workflows, and exportable artifacts.

**MANDATORY READ — `examples/`**: 20 self-contained HTML files, one per artifact type. Read the matching example before producing any artifact. They are the ground truth for structure, aesthetic, and interactivity level. The catalog table below maps task → filename.

## Core stance

A playground is not a mini-SaaS. It is a shipped interaction:
- browser as runtime
- synthetic/local data by default
- no auth/backend/API keys unless explicitly required
- one clear job per artifact
- polished enough to feel intentional

When the request is content-heavy, first ask: "Would interaction make the decision clearer?" If no, use Markdown. If yes, build HTML.

---

## Stack — always use these three

Every artifact uses this exact stack. No exceptions without explicit user override.

### 1. Tailwind CSS v4 (CDN)

```html
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
```

Configure custom tokens inline when needed:
```html
<style type="text/tailwindcss">
  @theme {
    --color-ivory: #FAF9F5;
    --color-slate: #141413;
    --color-clay: #D97757;
    --color-oat: #E3DACC;
    --color-olive: #788C5D;
  }
</style>
```

Use Tailwind classes for **all** layout, spacing, color, and typography. Only use `<style>` for things Tailwind cannot express: SVG strokes, scroll-snap, `@keyframes`, `view-transition-name`, `animation-timeline`.

### 2. Alpine.js v3

```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

Load order: Tailwind script → `<style type="text/tailwindcss">` → Alpine deferred. Always defer Alpine; it must load after the DOM.

Use Alpine for **all** reactive state. No manual `addEventListener` + `querySelector` DOM wiring when Alpine can do it with `x-model`, `x-show`, `x-on`, `x-bind`, `@click`, `$watch`.

For complex state, define a function and register it before Alpine boots:
```html
<script>
  function appState() {
    return {
      items: [],
      selected: null,
      init() { this.$watch('selected', v => console.log(v)); },
    };
  }
</script>
<div x-data="appState()">…</div>
```

### 3. Native browser APIs — reach for these before any library

| Feature | API | When |
|---|---|---|
| Overlays, menus, panels | Popover API (`popover` attr + `popovertarget`) | Non-blocking contextual content |
| Modal confirmation | `<dialog>` + `.showModal()` | Blocking decision |
| Animated state change | `document.startViewTransition()` | Panel swap, route-like nav, before/after reveal |
| Clipboard | `navigator.clipboard.writeText()` | Copy prompt, export |
| Persistence | `localStorage` | User saves state across sessions |
| URL state | `URLSearchParams` | Shareable filtered views |
| Disclosure | `<details><summary>` | Low-stakes collapse |

Always wrap `startViewTransition` and `showPopover` in feature checks:
```js
const update = () => { /* state change */ };
document.startViewTransition ? document.startViewTransition(update) : update();
```

**MANDATORY READ — `references/browser-apis.md`**: Popover API patterns, View Transitions, x-transition vs startViewTransition decision guide, CSS scroll-driven animations, and motion rules.

**MANDATORY READ — `references/alpine-plugins.md`**: Sort (drag-to-reorder), Persist, dark mode toggle, Collapse, Anchor, and common Alpine patterns (copy to clipboard, live filter).

---

## Artifact catalog — route to the right type

| Task | Type | Example |
|---|---|---|
| Compare N approaches side-by-side | Comparison / exploration | `01-exploration-code-approaches.html` |
| Explore design directions visually | Design explorer | `02-exploration-visual-designs.html` |
| Implementation plan with milestones | Spec / planning doc | `16-implementation-plan.html` |
| Annotated PR / diff review | Code review | `03-code-review-pr.html` |
| PR writeup for reviewers | Code review writeup | `17-pr-writeup.html` |
| Module map / dependency graph | Architecture diagram | `04-code-understanding.html` |
| Design system tokens showcase | Design tokens | `05-design-system.html` |
| Component variants contact sheet | Component explorer | `06-component-variants.html` |
| Animation / easing tuner | Interactive playground | `07-prototype-animation.html` |
| Clickable flow / prototype | Clickable prototype | `08-prototype-interaction.html` |
| SVG figures / diagrams | SVG diagram | `10-svg-illustrations.html` |
| Flowchart / deploy pipeline | Annotated flowchart | `13-flowchart-diagram.html` |
| Slide deck / presentation | Slideshow deck | `09-slide-deck.html` |
| Feature explainer with collapsibles | Research explainer | `14-research-feature-explainer.html` |
| Concept explainer with live demo | Concept explainer | `15-research-concept-explainer.html` |
| Weekly status / what shipped | Status report | `11-status-report.html` |
| Incident timeline / post-mortem | Incident report | `12-incident-report.html` |
| Ticket triage / drag-to-bucket | Throwaway editor | `18-editor-triage-board.html` |
| Feature flag config editor | Throwaway editor | `19-editor-feature-flags.html` |
| Prompt tuner / template editor | Throwaway editor | `20-editor-prompt-tuner.html` |
| Filterable table / log viewer | Data explorer | — |
| Database schema / ERD | ERD explorer | — |
| Branching concept map | Mind map | — |
| Gantt / roadmap / timeline | Timeline roadmap | — |
| Brainstorm N-variant grid | Brainstorm grid | — |
| Local-first single-user app (SQLite, OPFS, no backend) | Local-first app | `templates/local-first-app/` |

**MANDATORY READ — `references/html-artifacts.md`**: per-type guidance (slide decks, throwaway editors, diagrams, data explorers, local-first app, etc.), design system tokens, scannability rules, and GitHub Pages delivery.

**MANDATORY READ — `references/local-first.md`**: persistence ladder, JS-first stance, Worker architecture, AI integration ladder, and export/import pattern. Read before adding persistence, import/export, or model integration to any artifact.

---

## HTML output foundation

These rules apply to every artifact. Nothing below is optional.

**Write a real `.html` file — never inline-render in chat.** Every artifact is a file on disk (`<topic>-<kind>.html`), opened in a browser. Not a fenced ` ```html ` block, not a canvas widget. Inline rendering loses clipboard/network access, breaks dark-mode themes, and can't run the submit handler.

**Self-contained.** No build step, no npm, no external runtime. Tailwind CDN, Alpine CDN, and Google Fonts are the only external loads permitted. Everything else — JS logic, SVG, data — inline.

**Mobile-responsive.** Use Tailwind responsive prefixes (`sm:`, `md:`, `lg:`). Collapse to single column under `sm`. Artifacts get shared on phones during incidents and commutes.

**Semantic HTML.** Code in `<pre><code>`. Tables in `<table>`. Diagrams as inline `<svg>`. The reader must be able to copy any value or label out of the artifact.

**No `innerHTML` from variables.** Use `textContent` for text, Alpine's `x-text` for reactive text, and `document.createElement` + `appendChild` for dynamic structure. `innerHTML` from a variable is an XSS vector and trips security hooks in Claude Code and other harnesses. Static literal markup is fine.

**SVG text doesn't wrap — size the shape to the label.** For variable-length labels use `<foreignObject width="W" height="H">` with an HTML `<div>` inside. Plain `<text>` only for short fixed labels.

**Deliberate aesthetic — not the generic AI look.** No default purple gradient + Inter + three centered feature cards. Match visual direction to domain: utilitarian for ops artifacts, editorial for reports, engineering for diagrams. Pick a type pairing and commit.

**Accessible by default.** Body text meets WCAG AA contrast. Interactive controls keyboard-reachable with visible focus rings (`focus-visible:ring-2`). Status conveyed by shape/label, not color alone.

**Print-readable.** `Cmd+P` should produce something usable. Use Tailwind's `print:` variant to handle dark backgrounds and clipped content.

**Timestamp in footer** for any artifact someone might revisit — specs, diagrams, reports, roadmaps. One-shot editors and ephemeral playgrounds can omit.

**Filename is the artifact name.** `<topic>-<kind>.html`. Not `output.html`.

**Never start blank.** Pre-populate with sensible defaults, presets, or sample data. Empty playgrounds feel broken.

---

## Submit pipeline (server or clipboard)

For interactive artifacts whose value is what the user submits back.

| Mode | When |
|---|---|
| **Server** | Local Claude Code session with shell access. Run `html-skills-listen` skill first; it returns a URL. Inject `window.__CLAUDE_SUBMIT_URL__ = '<url>'` into the artifact. |
| **Clipboard** | Cloud / web / sandboxed, or whenever server mode isn't available. Always works; user pastes JSON back. |

### Submit handler pattern

Inline the submit handler and call `submitToClaude(payload)`:
```html
<button @click="submit()" class="px-6 py-3 bg-slate-900 text-white rounded-full font-mono text-sm">
  Submit to Claude
</button>
<script>
  // window.__CLAUDE_SUBMIT_URL__ = '…'; // server mode only
  function submitToClaude(payload) {
    const json = JSON.stringify(payload, null, 2);
    if (window.__CLAUDE_SUBMIT_URL__) {
      fetch(window.__CLAUDE_SUBMIT_URL__, { method: 'POST', body: json }).catch(() => fallback(json));
    } else {
      fallback(json);
    }
    function fallback(text) { navigator.clipboard.writeText(text); }
  }
</script>
```

Standard payload envelope — use the same shape across all types:
```json
{
  "skill": "playground",
  "kind":  "kanban-reorder",
  "data":  { },
  "version": 1
}
```

One Submit button per artifact. Don't fork into "Copy as JSON" and "Copy as prompt" — the envelope IS the export.

---

## AI workflow pattern

For org-safe AI labs, default to copy/paste mode:
- app provides prompt
- participant uses approved AI tool
- participant pastes result back
- app helps compare, score, or export

Do not call model APIs from static pages unless the user explicitly accepts API-key, auth, CORS, and data-handling implications.

## Output contract

Every playground makes these obvious in 5 seconds:
- what it is
- what to do first
- what changes when users interact
- what output they leave with

---

## NEVER

- **NEVER build a backend for a one-user or workshop artifact.**
  Use static files, browser storage, import/export, and copy/paste flows.

- **NEVER put API keys in browser code.**
  Use copy/paste with approved tools, or add a proper backend only after security review.

- **NEVER start with a blank control panel.**
  Provide sensible defaults, presets, or a sample scenario already loaded.

- **NEVER make interaction decorative.**
  Use interaction to compare, reveal, configure, copy, validate, or export.

- **NEVER hide the user's takeaway inside UI state.**
  Provide copy/export for prompts, decisions, summaries, or markdown.

- **NEVER write `<style>` blocks for things Tailwind covers.**
  Only `<style>` for SVG strokes, scroll-snap, `@keyframes`, `view-transition-name`, `animation-timeline`.

- **NEVER use `innerHTML` from a variable.**
  Use `x-text`, `textContent`, or `createElement` + `appendChild`.

- **NEVER write manual DOM wiring when Alpine can do it.**
  If you're writing `document.querySelector` + `addEventListener` for something Alpine's `x-model` / `@click` / `x-show` handles, stop and use Alpine.

- **NEVER inline-render in chat.** Always write the file.
