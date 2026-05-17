# Alpine Plugins & Patterns

## Alpine plugins

**All plugins must load before Alpine core (all deferred).**

### Sort — drag-to-reorder lists

```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/sort@3.x.x/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

Replaces hand-rolled HTML5 drag API. `x-sort` on the container, `x-sort:item` on children. Optional `x-sort:handle` restricts drag to a handle element. The handler receives `($item, $position)` for updating state.

```html
<ul x-data="{ items: ['Task A', 'Task B', 'Task C'] }"
    x-sort="reorder($item, $position)">
  <template x-for="(item, i) in items" :key="item">
    <li x-sort:item class="flex items-center gap-2 bg-white rounded-lg p-3 mb-2 shadow-sm">
      <span x-sort:handle class="cursor-grab text-gray-400">⠿</span>
      <span x-text="item"></span>
    </li>
  </template>
</ul>
```

```js
function appState() {
  return {
    items: ['Task A', 'Task B', 'Task C'],
    reorder(item, position) {
      this.items.splice(position, 0, this.items.splice(this.items.indexOf(item.dataset.key), 1)[0]);
    },
  };
}
```

Use for: kanban boards, reorderable lists, priority queues.

### Persist — localStorage without boilerplate

`$persist(value)` works inside `x-data` component scope and auto-syncs to localStorage.

```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/persist@3.x.x/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

```js
function appState() {
  return {
    tab: this.$persist('overview'),
    sidebarOpen: this.$persist(true),
  };
}
```

**Do not use `$persist` inside `Alpine.store()` or `alpine:init` listeners.** Both register via `alpine:init` internally; listener order from CDN is not guaranteed, so `Alpine.$persist` may not exist yet when your callback runs. Use manual `localStorage` read/write instead (see Dark Mode below).

### Dark mode toggle (canonical pattern)

No plugin needed. Manual `localStorage` avoids CDN timing issues entirely.

```html
<!-- In <head> -->
<script>
  document.addEventListener('alpine:init', () => {
    Alpine.store('theme', {
      dark: JSON.parse(localStorage.getItem('theme-dark') ?? 'null') ??
            window.matchMedia('(prefers-color-scheme: dark)').matches,
      toggle() {
        this.dark = !this.dark;
        localStorage.setItem('theme-dark', JSON.stringify(this.dark));
      }
    });
  });
</script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

`x-data` is required on `<html>` to open Alpine scope — `x-bind:class` is silently ignored without it:
```html
<html lang="en" x-data x-bind:class="{ dark: $store.theme.dark }">
```

Tailwind v4 — class-based dark variant (responds to `.dark` on `<html>`, not `prefers-color-scheme`):
```html
<style type="text/tailwindcss">
  @custom-variant dark (&:where(.dark, .dark *));
</style>
```

Toggle button:
```html
<button @click="$store.theme.toggle()"
        class="fixed top-3 right-3 z-50 w-8 h-8 flex items-center justify-center
               rounded-full bg-gray-100 dark:bg-gray-700 text-gray-500 dark:text-gray-300
               hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm"
        :title="$store.theme.dark ? 'Light mode' : 'Dark mode'"
        x-text="$store.theme.dark ? '☀' : '☾'">
</button>
```

### Collapse — smooth height animation

```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/collapse@3.x.x/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

`x-collapse` animates height from `0` to `auto` without fixed pixel values — better than `x-show` + manual CSS height transitions.

```html
<div x-data="{ open: false }">
  <button @click="open = !open"
          class="flex w-full items-center justify-between px-4 py-3 font-medium">
    Section title
    <span :class="open ? 'rotate-180' : ''" class="transition-transform duration-200">▾</span>
  </button>
  <div x-show="open" x-collapse class="px-4 pb-4 text-gray-600 text-sm">
    Collapsible content here.
  </div>
</div>
```

Use for: accordions, FAQ sections, expandable log entries.

### Anchor — floating element positioning

```html
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/anchor@3.x.x/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3/dist/cdn.min.js"></script>
```

Positions a floating element relative to a trigger via [Floating UI](https://floating-ui.com/). Replaces manual `getBoundingClientRect` + absolute positioning.

```html
<div x-data="{ open: false }">
  <button x-ref="btn" @click="open = !open"
          class="px-4 py-2 bg-slate-900 text-white rounded-full text-sm">
    Options
  </button>
  <div x-show="open"
       x-anchor.bottom-start="$refs.btn"
       x-transition
       class="absolute z-50 bg-white shadow-xl rounded-lg p-2 min-w-40">
    <button class="block w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded">Edit</button>
    <button class="block w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded">Delete</button>
  </div>
</div>
```

Positioning modifiers: `.top`, `.bottom`, `.left`, `.right`, `.top-start`, `.bottom-start`, `.top-end`, `.bottom-end`.

Use for: dropdowns, context menus, tooltips, autocomplete panels.

### Other official plugins

- **Focus** (`@alpinejs/focus`) — `x-trap="open"` traps keyboard focus inside a modal or panel.
- **Intersect** (`@alpinejs/intersect`) — `x-intersect` fires when an element enters/leaves the viewport. Use for lazy loading, scroll-triggered animations.
- **Mask** (`@alpinejs/mask`) — `x-mask` for input masking (phone numbers, credit cards, dates).
- **Morph** (`@alpinejs/morph`) — patches the DOM to match a new HTML string without full re-render.
- **Resize** (`@alpinejs/resize`) — `x-resize` fires when element dimensions change; replaces `ResizeObserver` boilerplate.

## Common Alpine patterns

### Copy to clipboard

```html
<button @click="copy()" x-text="copied ? 'Copied!' : 'Copy'" class="…"></button>
```
```js
{ copied: false,
  async copy() {
    await navigator.clipboard.writeText(this.output);
    this.copied = true;
    setTimeout(() => this.copied = false, 1500);
  }
}
```

### Live filter / search

```html
<div x-data="{ q: '', items: [...] }">
  <input x-model="q" placeholder="Filter…" class="w-full border rounded-lg px-3 py-2">
  <template x-for="item in items.filter(i => i.name.toLowerCase().includes(q.toLowerCase()))" :key="item.id">
    <div x-text="item.name"></div>
  </template>
</div>
```

For large lists, debounce `q` with `$watch` + `setTimeout` rather than filtering on every keystroke.
