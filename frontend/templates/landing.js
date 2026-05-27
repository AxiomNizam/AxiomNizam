/* =============================================
   AxiomNizam — Reimagined Landing Page JS
   Scroll reveals, counters, tabs, cursor glow, tilt
   ============================================= */
(function() {
    'use strict';

    // ---- Auth Helpers ----
    function readLandingCookie(name) {
        var prefix = name + '=';
        var parts = document.cookie.split(';');
        for (var i = 0; i < parts.length; i++) {
            var c = parts[i].trim();
            if (c.indexOf(prefix) === 0) {
                return decodeURIComponent(c.substring(prefix.length));
            }
        }
        return '';
    }

    function isAuthenticated() {
        return !!(localStorage.getItem('authToken') || readLandingCookie('authToken'));
    }

    window.requireAuth = function(targetPath) {
        if (isAuthenticated()) {
            window.location.href = targetPath;
        } else {
            window.location.href = '/signup';
        }
    };

    // ---- Cursor Glow ----
    var cursorGlow = document.getElementById('cursorGlow');
    if (cursorGlow) {
        var glowX = 0, glowY = 0, mouseX = 0, mouseY = 0;
        document.addEventListener('mousemove', function(e) {
            mouseX = e.clientX;
            mouseY = e.clientY;
            cursorGlow.classList.add('active');
        });
        document.addEventListener('mouseleave', function() {
            cursorGlow.classList.remove('active');
        });
        function animateGlow() {
            glowX += (mouseX - glowX) * 0.08;
            glowY += (mouseY - glowY) * 0.08;
            cursorGlow.style.left = glowX + 'px';
            cursorGlow.style.top = glowY + 'px';
            requestAnimationFrame(animateGlow);
        }
        animateGlow();
    }

    // ---- Scroll Reveal (IntersectionObserver) ----
    var revealElements = document.querySelectorAll('[data-reveal]');
    if (revealElements.length > 0) {
        var revealObserver = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    // Stagger siblings
                    var parent = entry.target.parentElement;
                    var siblings = parent ? parent.querySelectorAll('[data-reveal]') : [];
                    var delay = 0;
                    for (var i = 0; i < siblings.length; i++) {
                        if (siblings[i] === entry.target) {
                            delay = i * 60;
                            break;
                        }
                    }
                    setTimeout(function() {
                        entry.target.classList.add('revealed');
                    }, delay);
                    revealObserver.unobserve(entry.target);
                }
            });
        }, { threshold: 0.1, rootMargin: '0px 0px -40px 0px' });

        revealElements.forEach(function(el) {
            revealObserver.observe(el);
        });
    }

    // ---- Animated Counters ----
    var counterElements = document.querySelectorAll('[data-count]');
    var counterAnimated = false;

    function animateCounters() {
        if (counterAnimated) return;
        counterAnimated = true;

        counterElements.forEach(function(el) {
            var target = parseInt(el.getAttribute('data-count'), 10);
            if (isNaN(target)) return;
            var duration = 2000;
            var startTime = null;

            function step(timestamp) {
                if (!startTime) startTime = timestamp;
                var progress = Math.min((timestamp - startTime) / duration, 1);
                // Ease out cubic
                var eased = 1 - Math.pow(1 - progress, 3);
                el.textContent = Math.floor(eased * target);
                if (progress < 1) {
                    requestAnimationFrame(step);
                } else {
                    el.textContent = target;
                }
            }
            requestAnimationFrame(step);
        });
    }

    // Trigger counters when hero stats come into view
    var heroStats = document.querySelector('.hero__stats');
    if (heroStats) {
        var statsObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                animateCounters();
                statsObserver.unobserve(heroStats);
            }
        }, { threshold: 0.5 });
        statsObserver.observe(heroStats);
    }

    // ---- 3D Tilt Effect on Cards ----
    var tiltCards = document.querySelectorAll('[data-tilt]');
    tiltCards.forEach(function(card) {
        card.addEventListener('mousemove', function(e) {
            var rect = card.getBoundingClientRect();
            var x = e.clientX - rect.left;
            var y = e.clientY - rect.top;
            var centerX = rect.width / 2;
            var centerY = rect.height / 2;
            var rotateX = ((y - centerY) / centerY) * -4;
            var rotateY = ((x - centerX) / centerX) * 4;
            card.style.transform = 'perspective(800px) rotateX(' + rotateX + 'deg) rotateY(' + rotateY + 'deg) translateY(-4px)';
        });
        card.addEventListener('mouseleave', function() {
            card.style.transform = '';
        });
    });

    // ---- Deep Feature Tabs ----
    var deepTabs = document.querySelectorAll('.deep__tab');
    var deepPanels = document.querySelectorAll('.deep__panel');

    deepTabs.forEach(function(tab) {
        tab.addEventListener('click', function() {
            var targetId = 'panel-' + tab.getAttribute('data-tab');

            deepTabs.forEach(function(t) { t.classList.remove('active'); });
            deepPanels.forEach(function(p) { p.classList.remove('active'); });

            tab.classList.add('active');
            var targetPanel = document.getElementById(targetId);
            if (targetPanel) {
                targetPanel.classList.add('active');
            }
        });
    });

    // ---- Hero Platform Status ----
    var statusEl = document.getElementById('heroStatus');
    if (statusEl) {
        var backendURL = (document.getElementById('backendURL') || {}).textContent || '';
        backendURL = backendURL.trim();
        if (!backendURL) {
            backendURL = window.BACKEND_URL || 'http://localhost:8000';
        }
        // Remove trailing slash
        if (backendURL.length > 1 && backendURL.charAt(backendURL.length - 1) === '/') {
            backendURL = backendURL.slice(0, -1);
        }

        fetch(backendURL + '/health', { signal: AbortSignal.timeout ? AbortSignal.timeout(5000) : undefined })
            .then(function(r) { return r.json(); })
            .then(function() {
                statusEl.innerHTML = '<span class="pulse-dot pulse-dot--ok"></span>';
            })
            .catch(function() {
                statusEl.innerHTML = '<span class="pulse-dot" style="background:#ef4444"></span>';
            });
    }

    // ---- Keyboard shortcut: / to focus search ----
    document.addEventListener('keydown', function(e) {
        if (e.key === '/' && !e.ctrlKey && !e.metaKey && !e.altKey) {
            var tag = (document.activeElement || {}).tagName;
            if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;
            var searchInput = document.getElementById('featureSearch');
            if (searchInput) {
                e.preventDefault();
                searchInput.focus();
            }
        }
    });

})();
