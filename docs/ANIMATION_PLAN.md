# Landing Page Animation Plan

> **AxiomNizam Public Landing Page — Animation Enhancement Roadmap**
> Created: 2026-05-30 | Branch: `miraz-ui`
> Scope: Landing page only (`/` route, public-facing, no auth required)

---

## Overview

This document tracks all planned animations for the AxiomNizam **public landing page**. All work is scoped to the three landing page files — no dashboard or authenticated pages are affected.

### Files In Scope

| File | Role |
|------|------|
| `frontend/templates/public-dashboard.html` | Landing page HTML structure |
| `frontend/templates/landing.js` | Canvas animations, scroll effects, mouse interactions |
| `frontend/templates/landing.css` | CSS transitions, keyframes, hover effects |

### Sections Covered

| Section | ID | Current State |
|---------|-----|---------------|
| Hero | `#hero` | Animated orbs, stats, CTA |
| Bento Grid (Platform Capabilities) | `#bento` | 9 cards with canvas hover + demo data |
| Deep Dive (Tabbed Features) | `#deep-features` | 5 tabs with staggered animations |
| API Lifecycle | `#apiLifecycle` | Pipeline, node details, proximity effects |
| CLI Terminal | `#cli` | Interactive terminal |
| Architecture Diagram | `#arch` | SVG diagram |
| CTA | `#cta` | Final call-to-action |

---

## Completed Animations

| Animation | Section | Status | Files |
|-----------|---------|--------|-------|
| Canvas hover effects (API, Storage, CDC, Scanner, Conductor) | Bento Grid | Done | `landing.js`, `landing.css` |
| Canvas hover effects (Analytics, GIS, IAM, Network Intel) | Bento Grid | Done | `landing.js`, `landing.css` |
| Bento card demo data (all 9 cards) | Bento Grid | Done | `public-dashboard.html`, `landing.css` |
| Deep Dive staggered tab animation | Deep Dive | Done | `landing.js`, `landing.css` |
| Deep Dive scroll-triggered first appearance | Deep Dive | Done | `landing.js`, `landing.css` |
| API Lifecycle node hover detail popups | API Lifecycle | Done | `public-dashboard.html`, `landing.css` |
| API Lifecycle node 3D tilt on hover | API Lifecycle | Done | `landing.js`, `landing.css` |
| API Lifecycle request/response proximity animation | API Lifecycle | Done | `landing.js`, `landing.css` |
| API Lifecycle node active pulse | API Lifecycle | Done | `landing.css` |
| Cursor trail particles + glow + ring | Global | Done | `landing.js`, `landing.css` |
| Magnetic buttons | Hero CTA, Buttons | Done | `landing.js` |
| Scroll progress bar | Global (top) | Done | `landing.js`, `landing.css` |
| Floating badge animation | Bento tags | Done | `landing.css` |
| Text gradient animation | Section titles | Done | `landing.css` |
| Parallax on section headers | Bento, Deep, API, Arch | Done | `landing.js`, `landing.css` |
| Text scramble effect | Section eyebrows | Done | `landing.js` |
| Enhanced reveal types | All sections | Done | `landing.js`, `landing.css` |
| Card tilt with glare | Bento cards | Done | `landing.js`, `landing.css` |
| Data viz bar animation | Analytics demo | Done | `landing.js`, `landing.css` |
| CLI typing auto-play | CLI section | Done | `landing.js` |
| Toast notifications | Bento, API Lifecycle | Done | `landing.js`, `landing.css` |
| Horizontal scroll tabs | Deep Dive (mobile) | Done | `landing.css` |
| SVG icon draw-on animation | Bento cards | Done | `landing.css` |
| Interactive globe | GIS demo | Done | `landing.css` |
| Gradient mesh | Hero, sections | Done | `landing.css` |
| Section blur on scroll | All sections | Done | `landing.js`, `landing.css` |
| Scroll-linked timeline | API Lifecycle | Done | `landing.js`, `landing.css` |
| Hover card expand | Deep Dive items | Done | `landing.css` |
| 3D card stack | Architecture cards | Done | `landing.css` |

---

## Phase 1: Micro-Interactions (1-2 days) — COMPLETE

### 1.1 Cursor Trail Particles
- **What:** Colored particles follow the cursor across the entire page
- **Where:** Global
- **How:** Canvas overlay, spawn particles on mousemove, fade out over time
- **Complexity:** Low
- **Status:** Done (was already implemented)

### 1.2 Magnetic Buttons
- **What:** CTA buttons subtly follow cursor when mouse is nearby
- **Where:** Hero CTA, section CTAs
- **How:** Calculate distance, apply subtle transform toward cursor
- **Complexity:** Low
- **Status:** Done
- **What:** CTA buttons subtly follow cursor when mouse is nearby
- **Where:** Hero CTA, section CTAs
- **How:** Calculate distance, apply subtle transform toward cursor
- **Complexity:** Low

### 1.3 Scroll Progress Bar
- **What:** Thin animated bar at top showing page scroll progress
- **Where:** Fixed at top of viewport
- **How:** Track scroll position, update width of fixed bar
- **Complexity:** Low
- **Status:** Done

### 1.4 Floating Badge Animation
- **What:** Tags/badges float with subtle bob animation
- **Where:** Feature tags, bento card tags
- **How:** CSS keyframe with translateY oscillation
- **Complexity:** Low
- **Status:** Done

### 1.5 Text Gradient Animation
- **What:** Gradient text shifts colors smoothly
- **Where:** Section titles with `.text-gradient`
- **How:** CSS `background-position` animation on gradient text
- **Complexity:** Low
- **Status:** Done

---

## Phase 2: Scroll-Triggered Animations (2-3 days)

### 2.1 Animated Counters
- **What:** Numbers count up from 0 to target value when scrolling into view
- **Where:** Hero stats (111 modules, 1040 Go files, 244K+ Go lines)
- **How:** IntersectionObserver + requestAnimationFrame counter with easing
- **Complexity:** Medium
- **Status:** Done (enhanced with suffix support, number formatting, stagger, counting/counted states)

### 2.2 Section Parallax Layers
- **What:** Background elements move at different speeds during scroll
- **Where:** Hero orbs (mouse-reactive), section headers (scroll-based)
- **How:** Scroll event listener with transform calculations per layer
- **Complexity:** Medium
- **Status:** Done (enhanced with parallax on bento, deep, api-lifecycle, arch headers)

### 2.3 Text Scramble Effect
- **What:** Text characters scramble (random chars) before revealing final text
- **Where:** Section eyebrow labels (Platform Capabilities, Deep Dive, API Lifecycle, Architecture)
- **How:** Character-by-character reveal with random intermediate chars via IntersectionObserver
- **Complexity:** Medium
- **Status:** Done

### 2.4 Staggered Section Reveals
- **What:** Each section animates in with staggered child elements
- **Where:** All major sections
- **How:** IntersectionObserver + staggered CSS class application
- **Complexity:** Medium
- **Status:** Done (enhanced with reveal types: scale, slide-left, slide-right, rotate, blur)

### 2.5 Horizontal Scroll Gallery
- **What:** Features scroll horizontally with snap points
- **Where:** Deep Dive tabs on mobile
- **How:** CSS scroll-snap with horizontal overflow
- **Complexity:** Medium
- **Status:** Done

---

## Phase 3: Advanced Effects (3-5 days) — COMPLETE

### 3.1 Card Tilt with Glare
- **What:** Cards tilt on hover with a light reflection effect that follows mouse
- **Where:** Bento cards
- **How:** Mouse position tracking + perspective transform + radial gradient overlay
- **Complexity:** Medium
- **Status:** Done (added `.bento__card-glare` with mix-blend-mode overlay)

### 3.2 Morphing SVG Icons
- **What:** Icons animate with draw-on effect and hover pulse
- **Where:** Bento card icons, Deep Dive icons, Architecture icons
- **How:** SVG stroke-dasharray/dashoffset animation + drop-shadow pulse
- **Complexity:** Medium
- **Status:** Done

### 3.3 Interactive Globe
- **What:** CSS-based rotating globe with connection lines and pulsing dots
- **Where:** GIS Intelligence demo card
- **How:** CSS 3D transforms with rings, dots, and line animations
- **Complexity:** Medium
- **Status:** Done (CSS-based, no Three.js needed)

### 3.4 Data Visualization Motion
- **What:** Charts animate bars/lines/pies on load with easing
- **Where:** Analytics demo card
- **How:** IntersectionObserver + staggered height animation
- **Complexity:** Medium
- **Status:** Done

### 3.5 Typing Animation
- **What:** Terminal-style typing effect for code snippets
- **Where:** CLI section (auto-plays health check, api list, metrics show)
- **How:** Character-by-character reveal with command execution
- **Complexity:** Low-Medium
- **Status:** Done

### 3.6 Notification Toast Slide
- **What:** Toast notifications slide in from corner during demos
- **Where:** Bento section, API Lifecycle section
- **How:** CSS transform with JavaScript timer
- **Complexity:** Low
- **Status:** Done (`.toast-container`, `.toast`, `showToast()` function)

---

## Phase 4: Premium Effects (5-7 days) — COMPLETE

### 4.1 Animated Gradient Mesh
- **What:** Large animated gradient blobs floating in background
- **Where:** Hero section (5 orbs), bento/deep/api-lifecycle sections (mesh gradient)
- **How:** CSS radial gradients with keyframe position animation
- **Complexity:** Medium
- **Status:** Done

### 4.2 Page Transition Blur
- **What:** Sections blur/fade when scrolling between them
- **Where:** All major sections
- **How:** IntersectionObserver with blur/opacity transitions
- **Complexity:** Medium
- **Status:** Done

### 4.3 Scroll-Linked Timeline
- **What:** Progress bar syncs with API Lifecycle scroll position
- **Where:** API Lifecycle section
- **How:** Scroll position mapped to progress bar + stage highlights
- **Complexity:** Medium
- **Status:** Done (clickable stages, scroll-linked progress)

### 4.4 Hover Card Expand
- **What:** Cards expand to show more detail with smooth transition
- **Where:** Deep Dive items
- **How:** CSS max-height transition on hidden expand content
- **Complexity:** Medium
- **Status:** Done (.deep__item-expand with max-height animation)

### 4.5 Background Particles Field
- **What:** Subtle particle field that reacts to mouse movement
- **Where:** Hero section
- **How:** Canvas with particle physics, mouse repulsion/attraction
- **Complexity:** High
- **Status:** Done (was already implemented: 80 particles, repulsion, connections)

### 4.6 3D Card Stack
- **What:** Cards stack in 3D and fan out on hover
- **Where:** Architecture section cards
- **How:** CSS perspective + rotateX/Y transforms
- **Complexity:** Medium
- **Status:** Done

---

## Implementation Guidelines

### Performance Rules
- Use `requestAnimationFrame` for all JS animations
- Use `will-change: transform` for animated elements
- Use `transform` and `opacity` only (avoid layout-triggering properties)
- Use `IntersectionObserver` to pause off-screen animations
- Limit canvas particle count (max 100 active)
- Use `contain: layout` on animated containers

### File Organization (Landing Page Only)
- **CSS animations:** `frontend/templates/landing.css` (keyframes + transitions)
- **JS animations:** `frontend/templates/landing.js` (canvas, scroll, mouse tracking)
- **HTML structure:** `frontend/templates/public-dashboard.html` (data attributes, canvas elements)

> **Note:** All animation work stays within these 3 files. Dashboard pages (`dashboard.js`, `style.css`, `object-storage.js`, etc.) are out of scope.

### Naming Convention
- CSS keyframes: `@keyframes [section][Action]` (e.g., `deepItemIn`, `nodePulse`)
- CSS classes: `[section]__[element]--[state]` (e.g., `.proximity-active`)
- JS functions: `[section]Animation` (e.g., `analyticsAnimation`, `conductorAnimation`)

### Responsive Behavior
- **Desktop (1024px+):** All animations enabled
- **Tablet (768-1024px):** Disable hover canvases, keep scroll animations
- **Mobile (480-768px):** Disable hover canvases, reduce particle count
- **Tiny (<480px):** Disable all canvas animations, CSS-only transitions

---

## Priority Matrix

| Phase | Effort | Impact | Priority | Status |
|-------|--------|--------|----------|--------|
| Phase 1: Micro-Interactions | 2 days | High | **P0** | **Complete** |
| Phase 2: Scroll-Triggered | 3 days | High | **P0** | **Complete** |
| Phase 3: Advanced Effects | 5 days | Medium | **P1** | **Complete** |
| Phase 4: Premium Effects | 7 days | Medium | **P2** | **Complete** |
| **Total** | **17 days** | | | **All Done** |

---

## Dependencies

| Animation | Requires |
|-----------|----------|
| Interactive Globe | Three.js library or CSS 3D expertise |
| Data Visualization | SVG animation or Chart.js |
| Morphing Icons | SVG path data for each icon state |
| Background Particles | Canvas performance optimization |

---

*Last updated: 2026-05-31 (UTC+6)*
