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
    fetchLandingHealth();
    initScrollAnimations();
});
