# Browser APIs

## Native browser features

Reach for these before any library:

| Feature | API | When |
|---|---|---|
| Non-blocking overlay | Popover API | menus, detail panels, contextual info |
| Blocking modal | `<dialog>` + `.showModal()` | confirmation, decision required |
| Animated state change | `document.startViewTransition()` | panel swap, slide nav, before/after reveal |
| Clipboard | `navigator.clipboard.writeText()` | copy prompt, export, share |
| Disclosure | `<details><summary>` | low-stakes collapse |

Always wrap modern APIs in feature checks:
```js
const update = () => { /* state change */ };
document.startViewTransition ? document.startViewTransition(update) : update();
el.showPopover?.();
```

## Popover API patterns

### Basic
```html
<button popovertarget="panel" class="text-sm underline cursor-pointer">View details</button>
<div id="panel" popover class="bg-white rounded-xl shadow-xl p-6 max-w-sm">
  Content here
</div>
```

### With Alpine — dynamic content
```html
<button @click="openPanel(item)" class="…">Details</button>
<div x-ref="panel" popover
     @toggle="if ($event.newState === 'closed') selected = null"
     class="bg-white rounded-xl shadow-xl p-6 max-w-sm">
  <h3 x-text="selected?.title" class="font-serif text-lg mb-2"></h3>
  <p x-text="selected?.body" class="text-gray-600 text-sm"></p>
</div>
```
```js
openPanel(item) {
  const update = () => {
    this.selected = item;
    this.$nextTick(() => this.$refs.panel.showPopover?.());
  };
  document.startViewTransition ? document.startViewTransition(update) : update();
}
```

## View Transitions patterns

### Slide navigation
```js
next() {
  const go = () => this.slide = Math.min(this.slide + 1, this.total - 1);
  document.startViewTransition ? document.startViewTransition(go) : go();
}
```

### Tab / panel switching
```js
switchTab(name) {
  const go = () => this.tab = name;
  document.startViewTransition ? document.startViewTransition(go) : go();
}
```

### Named transition pairs (hero-style)
```html
<div style="view-transition-name: card-42" class="…">…</div>
```
```css
/* In <style> block — Tailwind can't express this */
::view-transition-old(card-42) { animation: slide-out 200ms ease; }
::view-transition-new(card-42) { animation: slide-in 200ms ease; }
```

## x-transition vs startViewTransition

Pick the right tool — they solve different problems:

| Situation | Use |
|---|---|
| Element enters/exits the DOM via `x-show` or `x-if` | `x-transition` |
| Two elements swap (one out, one in), or hero morph between positions | `startViewTransition` |
| Alpine state change that also toggles `x-show` elements | Both together |

### `x-transition` — element enters/exits

Handles opacity and scale automatically. No JS needed.

```html
<!-- Fade-in when visible, fade-out when hidden -->
<div x-show="open" x-transition>
  Content
</div>

<!-- Custom timing -->
<div
  x-show="open"
  x-transition:enter="transition ease-out duration-200"
  x-transition:enter-start="opacity-0 scale-95"
  x-transition:enter-end="opacity-100 scale-100"
  x-transition:leave="transition ease-in duration-150"
  x-transition:leave-start="opacity-100 scale-100"
  x-transition:leave-end="opacity-0 scale-95"
>
  Content
</div>
```

### `startViewTransition` — two elements swapping

Captures a before/after snapshot of the page and cross-fades them. Best for panel swaps, tab content, and named hero morphs.

```js
switchTab(name) {
  const go = () => this.tab = name;
  document.startViewTransition ? document.startViewTransition(go) : go();
}
```

```html
<!-- Named hero morph: same view-transition-name in both positions -->
<img style="view-transition-name: hero-img" src="thumb.jpg" class="w-24 h-24 rounded" />
<!-- (different DOM location, same name) -->
<img style="view-transition-name: hero-img" src="thumb.jpg" class="w-full h-64 object-cover" />
```

```css
/* In <style> — Tailwind can't express view-transition-name */
::view-transition-old(hero-img) { animation: fade-out 200ms ease; }
::view-transition-new(hero-img) { animation: fade-in 200ms ease; }
```

### Using both together

`startViewTransition` wraps the state mutation and snapshots the page before/after. `x-transition` on each `x-show` element handles its own enter/leave independently.

```js
openDetail(item) {
  const go = () => {
    this.selected = item;
    this.view = 'detail';
  };
  document.startViewTransition ? document.startViewTransition(go) : go();
}
```

```html
<div x-show="view === 'list'" x-transition>…list…</div>
<div x-show="view === 'detail'" x-transition>…detail…</div>
```

## CSS Scroll-Driven Animations

**Browser support:** Chrome/Edge/Opera (v115+) ✓ · Safari 18 ✓ · Firefox partial.

Runs on the compositor thread — zero jank, no JavaScript scroll listeners. Declare once in CSS; the browser handles everything.

### Timeline types

- `scroll()` — progress tied to a scrollable element (`scroll(root)` for the page).
- `view()` — progress tied to the element's position within the viewport (enters → exits).

### Cards fade+rise as they enter the viewport

```css
@keyframes fade-up {
  from { opacity: 0; translate: 0 24px; }
  to   { opacity: 1; translate: 0 0; }
}

.card {
  animation: fade-up linear both;
  animation-timeline: view();
  animation-range: entry 0% entry 60%;
}
```

### Reading progress bar

```css
.progress-bar {
  position: fixed; top: 0; left: 0;
  height: 3px; background: var(--color-clay);
  width: 0%;
  animation: grow-x linear;
  animation-timeline: scroll(root);
}

@keyframes grow-x { to { width: 100%; } }
```

### Sticky header shrinks on scroll

```css
.header {
  animation: shrink linear;
  animation-timeline: scroll(root);
  animation-range: 0px 200px;
}

@keyframes shrink {
  to { padding-block: 0.5rem; font-size: 0.875rem; }
}
```

### Where these go

Put scroll-driven animation CSS in `<style type="text/tailwindcss">` — Tailwind cannot express `animation-timeline` or `animation-range`.

### Progressive enhancement

Never gate visibility on animation completion. Content must be readable if scroll-driven animations are absent.

```css
@supports (animation-timeline: scroll()) {
  .card { opacity: 0; } /* only hide if supported */
}
```

## Motion rules

- Animate state changes that would otherwise feel abrupt.
- Keep transitions under ~200ms for utility UI; ~300ms for narrative (deck slides).
- `x-transition` on Alpine `x-show` adds a default fade — use it freely.
- Respect `prefers-reduced-motion`: Tailwind's `motion-safe:` and `motion-reduce:` variants.
- Never make content wait for animation. Transitions are feedback, not gates.
