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
- **Where:** Hero stats (111 modules, 244K lines, etc.), section stats
- **How:** IntersectionObserver + requestAnimationFrame counter
- **Complexity:** Medium

### 2.2 Section Parallax Layers
- **What:** Background elements move at different speeds during scroll
- **Where:** Hero orbs, section backgrounds
- **How:** Scroll event listener with transform calculations per layer
- **Complexity:** Medium

### 2.3 Text Scramble Effect
- **What:** Text characters scramble (random chars) before revealing final text
- **Where:** Hero title, section titles
- **How:** Character-by-character reveal with random intermediate chars
- **Complexity:** Medium

### 2.4 Staggered Section Reveals
- **What:** Each section animates in with staggered child elements
- **Where:** All major sections
- **How:** IntersectionObserver + staggered CSS class application
- **Complexity:** Medium

### 2.5 Horizontal Scroll Gallery
- **What:** Features scroll horizontally with snap points
- **Where:** Mobile feature showcase
- **How:** CSS scroll-snap with horizontal overflow
- **Complexity:** Medium

---

## Phase 3: Advanced Effects (3-5 days)

### 3.1 Card Tilt with Glare
- **What:** Cards tilt on hover with a light reflection effect that follows mouse
- **Where:** Bento cards, Deep Dive items
- **How:** Mouse position tracking + perspective transform + radial gradient overlay
- **Complexity:** Medium

### 3.2 Morphing SVG Icons
- **What:** Icons animate between states on hover (e.g., lock to unlock)
- **Where:** Feature icons, section icons
- **How:** SVG path morphing with `d` attribute animation
- **Complexity:** Medium

### 3.3 Interactive Globe
- **What:** 3D rotating globe with connection lines between cities
- **Where:** GIS Intelligence section
- **How:** Three.js or CSS 3D transforms with SVG paths
- **Complexity:** High

### 3.4 Data Visualization Motion
- **What:** Charts animate bars/lines/pies on load with easing
- **Where:** Analytics demo, dashboard previews
- **How:** SVG animation or Canvas drawing with requestAnimationFrame
- **Complexity:** Medium

### 3.5 Typing Animation
- **What:** Terminal-style typing effect for code snippets
- **Where:** CLI section, API demo
- **How:** Character-by-character reveal with cursor blink
- **Complexity:** Low-Medium

### 3.6 Notification Toast Slide
- **What:** Toast notifications slide in from corner during demos
- **Where:** Demo sections (scanner, CDC, conductor)
- **How:** CSS transform with JavaScript timer
- **Complexity:** Low

---

## Phase 4: Premium Effects (5-7 days)

### 4.1 Animated Gradient Mesh
- **What:** Large animated gradient blobs floating in background
- **Where:** Hero section, section backgrounds
- **How:** CSS radial gradients with keyframe position animation
- **Complexity:** Medium

### 4.2 Page Transition Blur
- **What:** Sections blur/fade when scrolling between them
- **Where:** Between major sections
- **How:** Scroll-based filter: blur() calculation
- **Complexity:** Medium

### 4.3 Scroll-Linked Timeline
- **What:** Progress bar syncs with API Lifecycle animation
- **Where:** API Lifecycle section
- **How:** Scroll position mapped to animation progress
- **Complexity:** Medium

### 4.4 Hover Card Expand
- **What:** Cards expand to show more detail with smooth transition
- **Where:** Deep Dive items, feature cards
- **How:** CSS grid-template-rows animation or height transition
- **Complexity:** Medium

### 4.5 Background Particles Field
- **What:** Subtle particle field that reacts to mouse movement
- **Where:** Hero section
- **How:** Canvas with particle physics, mouse repulsion/attraction
- **Complexity:** High

### 4.6 3D Card Stack
- **What:** Cards stack in 3D and fan out on hover
- **Where:** Feature showcase
- **How:** CSS perspective + rotateX/Y transforms
- **Complexity:** Medium

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

| Phase | Effort | Impact | Priority |
|-------|--------|--------|----------|
| Phase 1: Micro-Interactions | 2 days | High | **P0** |
| Phase 2: Scroll-Triggered | 3 days | High | **P0** |
| Phase 3: Advanced Effects | 5 days | Medium | **P1** |
| Phase 4: Premium Effects | 7 days | Medium | **P2** |
| **Total** | **17 days** | | |

---

## Dependencies

| Animation | Requires |
|-----------|----------|
| Interactive Globe | Three.js library or CSS 3D expertise |
| Data Visualization | SVG animation or Chart.js |
| Morphing Icons | SVG path data for each icon state |
| Background Particles | Canvas performance optimization |

---

*Last updated: 2026-05-30 (UTC+6)*
