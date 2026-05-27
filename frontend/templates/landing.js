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

    // ============================================================
    // COMMAND PALETTE SEARCH
    // ============================================================
    var searchOverlay = document.getElementById('searchOverlay');
    var searchBackdrop = document.getElementById('searchBackdrop');
    var searchModal = document.getElementById('searchModal');
    var searchInput = document.getElementById('searchInput');
    var searchResults = document.getElementById('searchResults');
    var searchEmpty = document.getElementById('searchEmpty');
    var searchTrigger = document.getElementById('searchTrigger');
    var activeIndex = -1;

    // Searchable items index
    var searchItems = [
        // Data Platform
        { title: 'API Builder', desc: 'Create REST APIs visually with auto-generated CRUD endpoints', icon: '🔌', url: '/admin', group: 'Data Platform', keywords: 'rest crud api mysql postgres mongodb oracle firebase openapi' },
        { title: 'Object Storage', desc: 'S3-compatible storage with buckets, encryption, and policies', icon: '🗄️', url: '/object-storage', group: 'Data Platform', keywords: 's3 bucket upload download minio storage' },
        { title: 'Data Sources', desc: 'Connect external databases, APIs, and data streams', icon: '🧬', url: '/admin', group: 'Data Platform', keywords: 'database connection pool external api' },
        { title: 'Dynamic Queries', desc: 'SQL queries across connected databases with schema introspection', icon: '⚡', url: '/admin', group: 'Data Platform', keywords: 'sql query database schema batch' },
        { title: 'GraphQL Studio', desc: 'Interactive playground with schema introspection', icon: '◈', url: '/admin', group: 'Data Platform', keywords: 'graphql query mutation subscription' },
        { title: 'SQL AI Assistant', desc: 'Natural language to SQL with AI-powered suggestions', icon: '🤖', url: '/admin', group: 'Data Platform', keywords: 'ai sql assistant natural language nlp' },
        { title: 'CSV Upload & Transform', desc: 'Upload CSV files and auto-generate dashboards and APIs', icon: '📄', url: '/admin', group: 'Data Platform', keywords: 'csv upload import transform data' },
        { title: 'Schema Registry', desc: 'Avro, Protobuf, and JSON schema management', icon: '📐', url: '/admin', group: 'Data Platform', keywords: 'schema avro protobuf json registry compatibility' },
        { title: 'Feature Store', desc: 'Feature group management with online/offline serving', icon: '🧬', url: '/admin', group: 'Data Platform', keywords: 'feature store ml ai machine learning serving' },

        // Real-time
        { title: 'CDC Pipelines', desc: 'Real-time Change Data Capture with streaming ingestion', icon: '🔄', url: '/cdc-etl', group: 'Real-time', keywords: 'cdc change data capture streaming pipeline etl' },
        { title: 'ETL Pipelines', desc: 'Extract-Transform-Load with 20+ connector catalog', icon: '🔀', url: '/cdc-etl', group: 'Real-time', keywords: 'etl extract transform load connector pipeline' },
        { title: 'Message Conductor', desc: 'RabbitMQ and Kafka broker management', icon: '🎵', url: '/conductor', group: 'Real-time', keywords: 'kafka rabbitmq message broker producer consumer' },
        { title: 'WebSocket Streaming', desc: 'Real-time data streaming via WebSockets and SSE', icon: '🌊', url: '/admin', group: 'Real-time', keywords: 'websocket sse streaming real-time event' },
        { title: 'Webhooks', desc: 'Event-driven integrations with delivery logging', icon: '🪝', url: '/admin', group: 'Real-time', keywords: 'webhook event callback integration' },
        { title: 'Event Bus', desc: 'Publish/subscribe event streaming with DLQ support', icon: '📡', url: '/admin', group: 'Real-time', keywords: 'event bus pubsub publish subscribe dlq' },
        { title: 'Distributed Tracing', desc: 'Trace ingestion, span correlation, service map', icon: '🔎', url: '/admin', group: 'Real-time', keywords: 'tracing trace span opentelemetry jaeger' },
        { title: 'OpenTelemetry', desc: 'OTLP trace propagation with structured access logging', icon: '📡', url: '/admin', group: 'Real-time', keywords: 'opentelemetry otlp observability' },

        // Analytics & Visualization
        { title: 'Analytics Dashboards', desc: 'Custom dashboards with charts, heatmaps, KPI widgets', icon: '📊', url: '/analytics', group: 'Analytics', keywords: 'analytics dashboard chart heatmap kpi widget' },
        { title: 'GIS Intelligence', desc: 'Multi-layer interactive maps with regions and markers', icon: '🌍', url: '/gis', group: 'Analytics', keywords: 'gis map geography geo spatial leaflet layers' },
        { title: 'Network Intelligence', desc: 'Log parsing, topology mapping, anomaly detection', icon: '🔍', url: '/netintel', group: 'Analytics', keywords: 'network netintel topology heatmap anomaly mac' },
        { title: 'Data Lineage', desc: 'Column-level tracking with impact analysis', icon: '🗺️', url: '/admin', group: 'Analytics', keywords: 'lineage column tracking upstream downstream impact' },
        { title: 'Version Lineage', desc: 'Resource version tracking with diff visualization', icon: '📌', url: '/admin', group: 'Analytics', keywords: 'version lineage diff history rollback' },

        // Security & IAM
        { title: 'IAM & RBAC', desc: 'Keycloak integration with multi-realm and OAuth2/OIDC', icon: '🔐', url: '/iam-admin', group: 'Security', keywords: 'iam rbac keycloak realm user role permission oauth oidc saml' },
        { title: '2FA / MFA (Gatekeeper)', desc: 'TOTP enrollment, backup codes, trusted devices', icon: '🛡️', url: '/two-factor', group: 'Security', keywords: '2fa mfa totp gatekeeper backup trusted device authenticator' },
        { title: 'CSRF Protection', desc: 'Double-submit cookie CSRF with token rotation', icon: '🛡️', url: '/admin', group: 'Security', keywords: 'csrf protection token cookie security' },
        { title: 'Security Headers', desc: 'HSTS, CSP, X-Frame-Options on every response', icon: '🔒', url: '/admin', group: 'Security', keywords: 'security headers hsts csp x-frame' },
        { title: 'Encryption & Certs', desc: 'AES encryption, key management, certificate lifecycle', icon: '🔑', url: '/admin', group: 'Security', keywords: 'encryption aes key certificate tls ssl' },
        { title: 'Governance Console', desc: 'Policy engine with access, admission, compliance policies', icon: '🏛️', url: '/governance', group: 'Security', keywords: 'governance policy compliance admission quota lifecycle' },
        { title: 'Audit & Compliance', desc: 'Complete audit trail with SOC2, HIPAA, PCI-DSS scoring', icon: '📋', url: '/admin', group: 'Security', keywords: 'audit compliance soc2 hipaa pci risk assessment' },

        // Infrastructure
        { title: 'Operations Center', desc: 'Incident management, alerting, and job scheduling', icon: '🛠️', url: '/admin', group: 'Infrastructure', keywords: 'operations center incident alert job schedule' },
        { title: 'Job Scheduler', desc: 'Cron-style scheduling with execution history', icon: '⏰', url: '/admin', group: 'Infrastructure', keywords: 'job scheduler cron task execution' },
        { title: 'Multi-Tenancy', desc: 'Tenant CRUD, member management, quota management', icon: '🏢', url: '/admin', group: 'Infrastructure', keywords: 'tenant multi-tenancy isolation quota member' },
        { title: 'Deployment Controller', desc: 'Blue-green and canary deployments with rollback', icon: '🚀', url: '/admin', group: 'Infrastructure', keywords: 'deployment blue-green canary rollback health' },
        { title: 'Trivy Scanner', desc: 'Vulnerability scanning for containers and repos', icon: '🏥', url: '/admin', group: 'Infrastructure', keywords: 'trivy vulnerability scan container image security' },
        { title: 'Autopilot', desc: 'Automated cluster health with dead server cleanup', icon: '🤖', url: '/admin', group: 'Infrastructure', keywords: 'autopilot cluster health quorum cleanup' },
        { title: 'SafeGate Scanner', desc: 'File scanning pipeline with MIME, SVG, macro detection', icon: '🛡️', url: '/object-storage', group: 'Infrastructure', keywords: 'safegate scanner mime svg macro antivirus archive' },

        // Advanced
        { title: 'Vector Search', desc: 'Semantic search with vector embeddings and KNN', icon: '🎨', url: '/admin', group: 'Advanced', keywords: 'vector search embedding knn semantic ai' },
        { title: 'K8s Extensions', desc: 'Admission webhooks, CRD registry, scheduler', icon: '🧠', url: '/admin', group: 'Advanced', keywords: 'kubernetes k8s admission webhook crd scheduler' },
        { title: 'ML Pipeline', desc: 'Model deployment, feature extraction, training', icon: '🤖', url: '/admin', group: 'Advanced', keywords: 'ml machine learning pipeline model training deploy' },
        { title: 'Prometheus Metrics', desc: 'Per-module metric collectors with /metrics endpoint', icon: '📊', url: '/admin', group: 'Advanced', keywords: 'prometheus metrics observability monitor' },

        // Pages
        { title: 'Admin Dashboard', desc: 'Full platform administration and management', icon: '🏠', url: '/admin', group: 'Pages', keywords: 'admin dashboard management' },
        { title: 'Sign Up', desc: 'Create a free account', icon: '📝', url: '/signup', group: 'Pages', keywords: 'signup register account create' },
        { title: 'Sign In', desc: 'Sign in to your account', icon: '🔑', url: '/login', group: 'Pages', keywords: 'login signin sign-in auth' },
    ];

    // Fuzzy match
    function fuzzyMatch(text, query) {
        var ti = 0, qi = 0;
        var score = 0;
        var matchPositions = [];
        text = text.toLowerCase();
        query = query.toLowerCase();

        while (ti < text.length && qi < query.length) {
            if (text[ti] === query[qi]) {
                score += (ti === 0 || text[ti - 1] === ' ' || text[ti - 1] === '-') ? 3 : 1;
                matchPositions.push(ti);
                qi++;
            }
            ti++;
        }
        return qi === query.length ? { score: score, positions: matchPositions } : null;
    }

    function highlightMatch(text, positions) {
        if (!positions || positions.length === 0) return escapeHTML(text);
        var result = '';
        var posSet = {};
        for (var i = 0; i < positions.length; i++) posSet[positions[i]] = true;
        var inMark = false;
        for (var j = 0; j < text.length; j++) {
            if (posSet[j] && !inMark) { result += '<mark>'; inMark = true; }
            else if (!posSet[j] && inMark) { result += '</mark>'; inMark = false; }
            result += escapeHTML(text[j]);
        }
        if (inMark) result += '</mark>';
        return result;
    }

    function escapeHTML(str) {
        return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    }

    function doSearch(query) {
        if (!query || query.length < 1) return [];

        var scored = [];
        for (var i = 0; i < searchItems.length; i++) {
            var item = searchItems[i];
            var titleMatch = fuzzyMatch(item.title, query);
            var kwMatch = fuzzyMatch(item.keywords, query);
            var descMatch = fuzzyMatch(item.desc, query);

            var bestScore = 0;
            var bestPositions = null;
            var matchField = '';

            if (titleMatch && titleMatch.score * 2 > bestScore) {
                bestScore = titleMatch.score * 2;
                bestPositions = titleMatch.positions;
                matchField = 'title';
            }
            if (kwMatch && kwMatch.score > bestScore) {
                bestScore = kwMatch.score;
                bestPositions = kwMatch.positions;
                matchField = 'keywords';
            }
            if (descMatch && descMatch.score * 0.5 > bestScore) {
                bestScore = descMatch.score * 0.5;
                bestPositions = descMatch.positions;
                matchField = 'desc';
            }

            if (bestScore > 0) {
                scored.push({
                    item: item,
                    score: bestScore,
                    positions: matchField === 'title' ? bestPositions : null
                });
            }
        }

        scored.sort(function(a, b) { return b.score - a.score; });
        return scored.slice(0, 12);
    }

    function renderResults(query) {
        var results = doSearch(query);

        if (!query || query.length < 1) {
            searchResults.innerHTML = '';
            searchResults.appendChild(searchEmpty);
            searchEmpty.classList.remove('hidden');
            activeIndex = -1;
            return;
        }

        searchEmpty.classList.add('hidden');

        if (results.length === 0) {
            searchResults.innerHTML = '<div class="search-no-results"><span>No results found</span>Try a different search term</div>';
            activeIndex = -1;
            return;
        }

        // Group by category
        var groups = {};
        var groupOrder = [];
        for (var i = 0; i < results.length; i++) {
            var groupName = results[i].item.group;
            if (!groups[groupName]) {
                groups[groupName] = [];
                groupOrder.push(groupName);
            }
            groups[groupName].push(results[i]);
        }

        var html = '';
        var idx = 0;
        for (var g = 0; g < groupOrder.length; g++) {
            var gName = groupOrder[g];
            html += '<div class="search-group"><div class="search-group__label">' + escapeHTML(gName) + '</div>';
            for (var r = 0; r < groups[gName].length; r++) {
                var result = groups[gName][r];
                var titleHTML = result.positions ? highlightMatch(result.item.title, result.positions) : escapeHTML(result.item.title);
                html += '<a class="search-result' + (idx === 0 ? ' active' : '') + '" href="' + result.item.url + '" data-index="' + idx + '" style="animation-delay:' + (idx * 40) + 'ms">'
                    + '<div class="search-result__icon">' + result.item.icon + '</div>'
                    + '<div class="search-result__text">'
                    + '<div class="search-result__title">' + titleHTML + '</div>'
                    + '<div class="search-result__desc">' + escapeHTML(result.item.desc) + '</div>'
                    + '</div>'
                    + '<div class="search-result__arrow"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><line x1="5" y1="12" x2="19" y2="12"/><polyline points="12 5 19 12 12 19"/></svg></div>'
                    + '</a>';
                idx++;
            }
            html += '</div>';
        }

        searchResults.innerHTML = html;
        activeIndex = 0;
    }

    function setActiveResult(index) {
        var items = searchResults.querySelectorAll('.search-result');
        if (items.length === 0) return;
        if (index < 0) index = items.length - 1;
        if (index >= items.length) index = 0;
        for (var i = 0; i < items.length; i++) {
            items[i].classList.toggle('active', i === index);
        }
        activeIndex = index;
        // Scroll into view
        var activeEl = searchResults.querySelector('.search-result.active');
        if (activeEl) {
            activeEl.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
        }
    }

    function openSearch() {
        searchOverlay.classList.add('open');
        document.body.style.overflow = 'hidden';
        setTimeout(function() { searchInput.focus(); }, 50);
    }

    function closeSearch() {
        searchOverlay.classList.remove('open');
        document.body.style.overflow = '';
        searchInput.value = '';
        renderResults('');
        activeIndex = -1;
    }

    // Event: trigger button
    if (searchTrigger) {
        searchTrigger.addEventListener('click', openSearch);
    }

    // Event: backdrop click
    if (searchBackdrop) {
        searchBackdrop.addEventListener('click', closeSearch);
    }

    // Event: input
    if (searchInput) {
        searchInput.addEventListener('input', function() {
            renderResults(searchInput.value.trim());
        });
    }

    // Event: keyboard
    document.addEventListener('keydown', function(e) {
        // Open: / or Cmd+K / Ctrl+K
        if ((e.key === '/' || (e.key === 'k' && (e.metaKey || e.ctrlKey))) && !searchOverlay.classList.contains('open')) {
            var tag = (document.activeElement || {}).tagName;
            if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return;
            e.preventDefault();
            openSearch();
            return;
        }

        if (!searchOverlay.classList.contains('open')) return;

        // Close: Escape
        if (e.key === 'Escape') {
            e.preventDefault();
            closeSearch();
            return;
        }

        // Navigate: Arrow keys
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            setActiveResult(activeIndex + 1);
            return;
        }
        if (e.key === 'ArrowUp') {
            e.preventDefault();
            setActiveResult(activeIndex - 1);
            return;
        }

        // Select: Enter
        if (e.key === 'Enter') {
            e.preventDefault();
            var active = searchResults.querySelector('.search-result.active');
            if (active) {
                var href = active.getAttribute('href');
                if (isAuthenticated() || href === '/signup' || href === '/login') {
                    window.location.href = href;
                } else {
                    window.location.href = '/signup';
                }
            }
            return;
        }
    });

    // Event: result hover/click
    if (searchResults) {
        searchResults.addEventListener('mouseover', function(e) {
            var result = e.target.closest('.search-result');
            if (result) {
                var idx = parseInt(result.getAttribute('data-index'), 10);
                if (!isNaN(idx)) setActiveResult(idx);
            }
        });
    }

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
