// AxiomNizam Landing Page Dashboard JS

const BACKEND_URL = (() => {
    if (typeof window.resolveBackendURL === 'function') {
        return window.resolveBackendURL();
    }
    const value = String(window.BACKEND_URL || '').trim();
    if (value) return value.endsWith('/') ? value.slice(0, -1) : value;
    return 'http://localhost:8000';
})();

function isAuthenticated() {
    return !!(localStorage.getItem('authToken') || readLandingCookie('authToken'));
}

function readLandingCookie(name) {
    const prefix = name + '=';
    const parts = document.cookie.split(';');
    for (let i = 0; i < parts.length; i++) {
        const item = parts[i].trim();
        if (item.startsWith(prefix)) return decodeURIComponent(item.substring(prefix.length));
    }
    return '';
}

const LANDING_SVG_ICONS = {
    category: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1.5"></rect><rect x="14" y="3" width="7" height="7" rx="1.5"></rect><rect x="3" y="14" width="7" height="7" rx="1.5"></rect><rect x="14" y="14" width="7" height="7" rx="1.5"></rect></svg>',
    feature: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13 2 4 14h6l-1 8 9-12h-6z"></path></svg>',
    database: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><ellipse cx="12" cy="5" rx="8" ry="3"></ellipse><path d="M4 5v14c0 1.66 3.58 3 8 3s8-1.34 8-3V5"></path><path d="M4 12c0 1.66 3.58 3 8 3s8-1.34 8-3"></path></svg>',
    architecture: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="7" height="7" rx="1.5"></rect><rect x="14" y="3" width="7" height="7" rx="1.5"></rect><rect x="8.5" y="14" width="7" height="7" rx="1.5"></rect><path d="M10 6.5h4"></path><path d="M12 10v4"></path></svg>'
};

function replaceLandingEmojiIcons() {
    const landingRoot = document.querySelector('.landing-page');
    if (!landingRoot) return;

    const iconTargets = [
        ['.category-icon', LANDING_SVG_ICONS.category],
        ['.feature-card-icon', LANDING_SVG_ICONS.feature],
        ['.db-icon', LANDING_SVG_ICONS.database],
        ['.arch-icon', LANDING_SVG_ICONS.architecture]
    ];

    iconTargets.forEach(function(pair) {
        const selector = pair[0];
        const svg = pair[1];
        landingRoot.querySelectorAll(selector).forEach(function(node) {
            node.innerHTML = svg;
            node.setAttribute('aria-hidden', 'true');
        });
    });
}

// Gate feature access behind auth
function requireAuth(targetPath) {
    if (isAuthenticated()) {
        window.location.href = targetPath;
    } else {
        window.location.href = '/signup';
    }
}

// Fetch health status for the hero section
function fetchLandingHealth() {
    const el = document.getElementById('statHealth');
    if (!el) return;

    fetch('/api/health', { mode: 'cors' })
        .then(function(r) {
            if (!r.ok) throw new Error('HTTP ' + r.status);
            return r.json();
        })
        .then(function(data) {
            const status = (data.status || 'unknown').toUpperCase();
            if (status === 'OK' || status === 'HEALTHY') {
                el.innerHTML = '<span class="stat-dot stat-dot-ok"></span> Online';
            } else {
                el.innerHTML = '<span class="stat-dot stat-dot-warn"></span> ' + status;
            }
        })
        .catch(function() {
            el.innerHTML = '<span class="stat-dot stat-dot-err"></span> Offline';
        });
}

// Animate feature cards on scroll
function initScrollAnimations() {
    const cards = document.querySelectorAll('.feature-card, .arch-card, .db-card');
    if (!('IntersectionObserver' in window)) {
        cards.forEach(function(c) { c.classList.add('visible'); });
        return;
    }
    const observer = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
            if (entry.isIntersecting) {
                entry.target.classList.add('visible');
                observer.unobserve(entry.target);
            }
        });
    }, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });
    cards.forEach(function(c) { observer.observe(c); });
}

window.addEventListener('DOMContentLoaded', function() {
    replaceLandingEmojiIcons();
    fetchLandingHealth();
    initScrollAnimations();
});
