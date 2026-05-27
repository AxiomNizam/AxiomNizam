/* =============================================
   AxiomNizam — Reimagined Landing Page JS
   Particles, parallax, cursor glow, 3D tilt,
   counters, tabs, magnetic hover, spotlight
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

    // ---- Particle System ----
    var canvas = document.getElementById('particleCanvas');
    if (canvas) {
        var ctx = canvas.getContext('2d');
        var particles = [];
        var particleCount = 80;
        var connectionDistance = 150;
        var mouseParticle = { x: -1000, y: -1000, radius: 150 };

        function resizeCanvas() {
            canvas.width = canvas.offsetWidth;
            canvas.height = canvas.offsetHeight;
        }
        resizeCanvas();
        window.addEventListener('resize', resizeCanvas);

        function Particle() {
            this.x = Math.random() * canvas.width;
            this.y = Math.random() * canvas.height;
            this.vx = (Math.random() - 0.5) * 0.5;
            this.vy = (Math.random() - 0.5) * 0.5;
            this.radius = Math.random() * 1.5 + 0.5;
            this.opacity = Math.random() * 0.5 + 0.2;
        }

        for (var i = 0; i < particleCount; i++) {
            particles.push(new Particle());
        }

        // Track mouse for particle repulsion
        document.addEventListener('mousemove', function(e) {
            var rect = canvas.getBoundingClientRect();
            mouseParticle.x = e.clientX - rect.left;
            mouseParticle.y = e.clientY - rect.top;
        });

        function animateParticles() {
            ctx.clearRect(0, 0, canvas.width, canvas.height);

            for (var i = 0; i < particles.length; i++) {
                var p = particles[i];

                // Mouse repulsion
                var dx = p.x - mouseParticle.x;
                var dy = p.y - mouseParticle.y;
                var dist = Math.sqrt(dx * dx + dy * dy);
                if (dist < mouseParticle.radius) {
                    var force = (mouseParticle.radius - dist) / mouseParticle.radius;
                    p.vx += (dx / dist) * force * 0.3;
                    p.vy += (dy / dist) * force * 0.3;
                }

                // Damping
                p.vx *= 0.99;
                p.vy *= 0.99;

                p.x += p.vx;
                p.y += p.vy;

                // Wrap around
                if (p.x < -10) p.x = canvas.width + 10;
                if (p.x > canvas.width + 10) p.x = -10;
                if (p.y < -10) p.y = canvas.height + 10;
                if (p.y > canvas.height + 10) p.y = -10;

                // Draw particle
                ctx.beginPath();
                ctx.arc(p.x, p.y, p.radius, 0, Math.PI * 2);
                ctx.fillStyle = 'rgba(52, 211, 153, ' + p.opacity + ')';
                ctx.fill();

                // Draw connections
                for (var j = i + 1; j < particles.length; j++) {
                    var p2 = particles[j];
                    var dx2 = p.x - p2.x;
                    var dy2 = p.y - p2.y;
                    var dist2 = Math.sqrt(dx2 * dx2 + dy2 * dy2);
                    if (dist2 < connectionDistance) {
                        var alpha = (1 - dist2 / connectionDistance) * 0.15;
                        ctx.beginPath();
                        ctx.moveTo(p.x, p.y);
                        ctx.lineTo(p2.x, p2.y);
                        ctx.strokeStyle = 'rgba(52, 211, 153, ' + alpha + ')';
                        ctx.lineWidth = 0.5;
                        ctx.stroke();
                    }
                }

                // Draw connection to mouse
                if (dist < mouseParticle.radius * 1.5) {
                    var alphaMouse = (1 - dist / (mouseParticle.radius * 1.5)) * 0.2;
                    ctx.beginPath();
                    ctx.moveTo(p.x, p.y);
                    ctx.lineTo(mouseParticle.x, mouseParticle.y);
                    ctx.strokeStyle = 'rgba(56, 189, 248, ' + alphaMouse + ')';
                    ctx.lineWidth = 0.8;
                    ctx.stroke();
                }
            }

            requestAnimationFrame(animateParticles);
        }
        animateParticles();
    }

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
            var rotateX = ((y - centerY) / centerY) * -5;
            var rotateY = ((x - centerX) / centerX) * 5;
            card.style.transform = 'perspective(800px) rotateX(' + rotateX + 'deg) rotateY(' + rotateY + 'deg) translateY(-6px)';

            // Move spotlight
            var spotlight = card.querySelector('.bento__card-spotlight');
            if (spotlight) {
                spotlight.style.left = x + 'px';
                spotlight.style.top = y + 'px';
            }
        });
        card.addEventListener('mouseleave', function() {
            card.style.transform = '';
        });
    });

    // ---- Magnetic Hover on Arch Cards ----
    var magneticCards = document.querySelectorAll('.arch__card');
    magneticCards.forEach(function(card) {
        card.setAttribute('data-magnetic', '');
        card.addEventListener('mousemove', function(e) {
            var rect = card.getBoundingClientRect();
            var x = e.clientX - rect.left;
            var y = e.clientY - rect.top;
            card.style.setProperty('--mouse-x', x + 'px');
            card.style.setProperty('--mouse-y', y + 'px');

            // Subtle magnetic pull
            var centerX = rect.width / 2;
            var centerY = rect.height / 2;
            var pullX = ((x - centerX) / centerX) * 4;
            var pullY = ((y - centerY) / centerY) * 4;
            card.style.transform = 'translateY(-6px) translate(' + pullX + 'px, ' + pullY + 'px)';
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

    // ---- Parallax on Scroll ----
    var parallaxElements = document.querySelectorAll('[data-parallax]');
    if (parallaxElements.length > 0) {
        var ticking = false;
        window.addEventListener('scroll', function() {
            if (!ticking) {
                requestAnimationFrame(function() {
                    var scrollY = window.pageYOffset;
                    parallaxElements.forEach(function(el) {
                        var speed = parseFloat(el.getAttribute('data-parallax')) || 0.3;
                        var rect = el.getBoundingClientRect();
                        var offset = (rect.top + scrollY) * speed;
                        el.style.transform = 'translateY(' + (scrollY * speed - offset) + 'px)';
                    });
                    ticking = false;
                });
                ticking = true;
            }
        });
    }

    // ---- Hero Parallax Orbs (mouse-reactive) ----
    var heroSection = document.getElementById('hero');
    var heroOrbs = document.querySelectorAll('.hero__orb');
    if (heroSection && heroOrbs.length > 0) {
        heroSection.addEventListener('mousemove', function(e) {
            var rect = heroSection.getBoundingClientRect();
            var x = (e.clientX - rect.left) / rect.width - 0.5;
            var y = (e.clientY - rect.top) / rect.height - 0.5;

            heroOrbs.forEach(function(orb, i) {
                var depth = (i + 1) * 8;
                orb.style.transform = 'translate(' + (x * depth) + 'px, ' + (y * depth) + 'px)';
            });
        });
    }

    // ---- Terminal Typing Effect ----
    var terminalBody = document.querySelector('.terminal__body');
    if (terminalBody) {
        var lines = terminalBody.querySelectorAll('.terminal__line');
        var termObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                lines.forEach(function(line, i) {
                    line.style.opacity = '0';
                    line.style.transform = 'translateX(-10px)';
                    line.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
                    setTimeout(function() {
                        line.style.opacity = '1';
                        line.style.transform = 'translateX(0)';
                    }, i * 120);
                });
                termObserver.unobserve(terminalBody);
            }
        }, { threshold: 0.3 });
        termObserver.observe(terminalBody);
    }

    // ---- Hero Platform Status ----
    var statusEl = document.getElementById('heroStatus');
    if (statusEl) {
        var backendURL = (document.getElementById('backendURL') || {}).textContent || '';
        backendURL = backendURL.trim();
        if (!backendURL) {
            backendURL = window.BACKEND_URL || 'http://localhost:8000';
        }
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

})();
