# Stack Reference

## Tailwind CSS v4 (CDN)

```html
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
```

Custom tokens via `@theme` in a `<style type="text/tailwindcss">` block:

```html
<style type="text/tailwindcss">
  @theme {
    --color-ivory: #FAF9F5;
    --color-slate-dark: #141413;
    --color-clay: #D97757;
    --color-oat: #E3DACC;
    --color-olive: #788C5D;
  }
</style>
```

Use Tailwind classes for everything. Only write `<style>` for: SVG stroke/fill attributes, scroll-snap, `@keyframes`, `view-transition-name`, `animation-timeline`.

Load order: Tailwind script → `<style type="text/tailwindcss">` → Alpine deferred.

## Alpine.js v3

```html
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

Must be deferred — loads after DOM. For complex state, define a function and register it before Alpine boots:

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

Key directives:

| Directive | Purpose |
|---|---|
| `x-data` | Declares reactive scope |
| `x-model` | Two-way bind input → state |
| `x-show` / `x-if` | Conditional rendering (`x-show` keeps DOM, `x-if` removes it) |
| `x-text` | Bind text content (prefer over `x-html` to avoid XSS) |
| `x-transition` | Auto-animate `x-show` changes |
| `x-for` | List rendering (always use `:key`) |
| `@click`, `@keydown` | Event handlers |
| `x-ref` | Template ref for direct DOM access |
| `$watch` | Side-effect on state change |
| `$nextTick` | Run after Alpine re-renders |

**Never use `innerHTML` from a variable.** Use `x-text` for reactive text, `createElement` + `appendChild` for dynamic structure.
