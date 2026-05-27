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
            // Save intended destination so login can redirect back
            localStorage.setItem('returnTo', targetPath);
            window.location.href = '/login';
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

    // SVG icon helpers for search results
    var svgIcon = function(path, vb) {
        return '<svg viewBox="' + (vb || '0 0 24 24') + '" fill="none" stroke="currentColor" stroke-width="1.5" style="width:18px;height:18px;flex-shrink:0">' + path + '</svg>';
    };
    var ICONS = {
        db: svgIcon('<ellipse cx="12" cy="5" rx="9" ry="3"/><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"/><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"/>'),
        api: svgIcon('<path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>'),
        bolt: svgIcon('<polygon points="13 2 3 14 12 14 11 22 21 10 12 10"/>'),
        query: svgIcon('<rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 9h6"/><path d="M9 13h6"/><path d="M9 17h4"/>'),
        file: svgIcon('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>'),
        shuffle: svgIcon('<polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/>'),
        layers: svgIcon('<rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/>'),
        wave: svgIcon('<path d="M22 12h-4l-3 9L9 3l-3 9H2"/>'),
        eye: svgIcon('<path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7z"/><circle cx="12" cy="12" r="3"/>'),
        link: svgIcon('<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>'),
        radio: svgIcon('<circle cx="12" cy="12" r="2"/><path d="M16.24 7.76a6 6 0 0 1 0 8.49"/><path d="M7.76 16.24a6 6 0 0 1 0-8.49"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14"/><path d="M4.93 19.07a10 10 0 0 1 0-14.14"/>'),
        search: svgIcon('<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>'),
        chart: svgIcon('<path d="M18 20V10"/><path d="M12 20V4"/><path d="M6 20v-6"/>'),
        globe: svgIcon('<circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>'),
        zap: svgIcon('<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>'),
        pin: svgIcon('<polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>'),
        shield: svgIcon('<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><polyline points="9 12 11 14 15 10"/>'),
        lock: svgIcon('<rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>'),
        key: svgIcon('<path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4"/>'),
        building: svgIcon('<path d="M3 21h18"/><path d="M5 21V7l8-4 8 4v14"/><path d="M9 21v-6h6v6"/><path d="M9 9h1"/><path d="M14 9h1"/>'),
        checklist: svgIcon('<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><path d="M9 15l2 2 4-4"/>'),
        server: svgIcon('<rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/>'),
        clock: svgIcon('<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>'),
        users: svgIcon('<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 0 0-3-3.87"/><path d="M16 3.13a4 4 0 0 1 0 7.75"/>'),
        send: svgIcon('<path d="M22 2L11 13"/><path d="M22 2l-7 20-4-9-9-4z"/>'),
        shieldAlert: svgIcon('<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M12 8v4"/><circle cx="12" cy="16" r="0.5" fill="currentColor"/>'),
        gear: svgIcon('<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/>'),
        home: svgIcon('<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><polyline points="9 22 9 12 15 12 15 22"/>'),
        userPlus: svgIcon('<path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/>'),
    };

    // Searchable items index
    var searchItems = [
        // Data Platform
        { title: 'API Builder', desc: 'Create REST APIs visually with auto-generated CRUD endpoints', icon: ICONS.api, url: '/admin', group: 'Data Platform', keywords: 'rest crud api mysql postgres mongodb oracle firebase openapi' },
        { title: 'Object Storage', desc: 'S3-compatible storage with buckets, encryption, and policies', icon: ICONS.server, url: '/object-storage', group: 'Data Platform', keywords: 's3 bucket upload download minio storage' },
        { title: 'Data Sources', desc: 'Connect external databases, APIs, and data streams', icon: ICONS.db, url: '/admin', group: 'Data Platform', keywords: 'database connection pool external api' },
        { title: 'Dynamic Queries', desc: 'SQL queries across connected databases with schema introspection', icon: ICONS.bolt, url: '/admin', group: 'Data Platform', keywords: 'sql query database schema batch' },
        { title: 'GraphQL Studio', desc: 'Interactive playground with schema introspection', icon: ICONS.layers, url: '/admin', group: 'Data Platform', keywords: 'graphql query mutation subscription' },
        { title: 'SQL AI Assistant', desc: 'Natural language to SQL with AI-powered suggestions', icon: ICONS.query, url: '/admin', group: 'Data Platform', keywords: 'ai sql assistant natural language nlp' },
        { title: 'CSV Upload & Transform', desc: 'Upload CSV files and auto-generate dashboards and APIs', icon: ICONS.file, url: '/admin', group: 'Data Platform', keywords: 'csv upload import transform data' },
        { title: 'Schema Registry', desc: 'Avro, Protobuf, and JSON schema management', icon: ICONS.server, url: '/admin', group: 'Data Platform', keywords: 'schema avro protobuf json registry compatibility' },
        { title: 'Feature Store', desc: 'Feature group management with online/offline serving', icon: ICONS.layers, url: '/admin', group: 'Data Platform', keywords: 'feature store ml ai machine learning serving' },

        // Real-time
        { title: 'CDC Pipelines', desc: 'Real-time Change Data Capture with streaming ingestion', icon: ICONS.radio, url: '/cdc-etl', group: 'Real-time', keywords: 'cdc change data capture streaming pipeline etl' },
        { title: 'ETL Pipelines', desc: 'Extract-Transform-Load with 20+ connector catalog', icon: ICONS.shuffle, url: '/cdc-etl', group: 'Real-time', keywords: 'etl extract transform load connector pipeline' },
        { title: 'Message Conductor', desc: 'RabbitMQ and Kafka broker management', icon: ICONS.wave, url: '/conductor', group: 'Real-time', keywords: 'kafka rabbitmq message broker producer consumer' },
        { title: 'WebSocket Streaming', desc: 'Real-time data streaming via WebSockets and SSE', icon: ICONS.eye, url: '/admin', group: 'Real-time', keywords: 'websocket sse streaming real-time event' },
        { title: 'Webhooks', desc: 'Event-driven integrations with delivery logging', icon: ICONS.link, url: '/admin', group: 'Real-time', keywords: 'webhook event callback integration' },
        { title: 'Event Bus', desc: 'Publish/subscribe event streaming with DLQ support', icon: ICONS.radio, url: '/admin', group: 'Real-time', keywords: 'event bus pubsub publish subscribe dlq' },
        { title: 'Distributed Tracing', desc: 'Trace ingestion, span correlation, service map', icon: ICONS.search, url: '/admin', group: 'Real-time', keywords: 'tracing trace span opentelemetry jaeger' },
        { title: 'OpenTelemetry', desc: 'OTLP trace propagation with structured access logging', icon: ICONS.layers, url: '/admin', group: 'Real-time', keywords: 'opentelemetry otlp observability' },

        // Analytics & Visualization
        { title: 'Analytics Dashboards', desc: 'Custom dashboards with charts, heatmaps, KPI widgets', icon: ICONS.chart, url: '/analytics', group: 'Analytics', keywords: 'analytics dashboard chart heatmap kpi widget' },
        { title: 'GIS Intelligence', desc: 'Multi-layer interactive maps with regions and markers', icon: ICONS.globe, url: '/gis', group: 'Analytics', keywords: 'gis map geography geo spatial leaflet layers' },
        { title: 'Network Intelligence', desc: 'Log parsing, topology mapping, anomaly detection', icon: ICONS.search, url: '/netintel', group: 'Analytics', keywords: 'network netintel topology heatmap anomaly mac' },
        { title: 'Data Lineage', desc: 'Column-level tracking with impact analysis', icon: ICONS.zap, url: '/admin', group: 'Analytics', keywords: 'lineage column tracking upstream downstream impact' },
        { title: 'Version Lineage', desc: 'Resource version tracking with diff visualization', icon: ICONS.pin, url: '/admin', group: 'Analytics', keywords: 'version lineage diff history rollback' },

        // Security & IAM
        { title: 'IAM & RBAC', desc: 'Keycloak integration with multi-realm and OAuth2/OIDC', icon: ICONS.lock, url: '/iam-admin', group: 'Security', keywords: 'iam rbac keycloak realm user role permission oauth oidc saml' },
        { title: '2FA / MFA (Gatekeeper)', desc: 'TOTP enrollment, backup codes, trusted devices', icon: ICONS.shield, url: '/two-factor', group: 'Security', keywords: '2fa mfa totp gatekeeper backup trusted device authenticator' },
        { title: 'CSRF Protection', desc: 'Double-submit cookie CSRF with token rotation', icon: ICONS.shieldAlert, url: '/admin', group: 'Security', keywords: 'csrf protection token cookie security' },
        { title: 'Security Headers', desc: 'HSTS, CSP, X-Frame-Options on every response', icon: ICONS.lock, url: '/admin', group: 'Security', keywords: 'security headers hsts csp x-frame' },
        { title: 'Encryption & Certs', desc: 'AES encryption, key management, certificate lifecycle', icon: ICONS.key, url: '/admin', group: 'Security', keywords: 'encryption aes key certificate tls ssl' },
        { title: 'Governance Console', desc: 'Policy engine with access, admission, compliance policies', icon: ICONS.building, url: '/governance', group: 'Security', keywords: 'governance policy compliance admission quota lifecycle' },
        { title: 'Audit & Compliance', desc: 'Complete audit trail with SOC2, HIPAA, PCI-DSS scoring', icon: ICONS.checklist, url: '/admin', group: 'Security', keywords: 'audit compliance soc2 hipaa pci risk assessment' },

        // Infrastructure
        { title: 'Operations Center', desc: 'Incident management, alerting, and job scheduling', icon: ICONS.server, url: '/admin', group: 'Infrastructure', keywords: 'operations center incident alert job schedule' },
        { title: 'Job Scheduler', desc: 'Cron-style scheduling with execution history', icon: ICONS.clock, url: '/admin', group: 'Infrastructure', keywords: 'job scheduler cron task execution' },
        { title: 'Multi-Tenancy', desc: 'Tenant CRUD, member management, quota management', icon: ICONS.users, url: '/admin', group: 'Infrastructure', keywords: 'tenant multi-tenancy isolation quota member' },
        { title: 'Deployment Controller', desc: 'Blue-green and canary deployments with rollback', icon: ICONS.send, url: '/admin', group: 'Infrastructure', keywords: 'deployment blue-green canary rollback health' },
        { title: 'Trivy Scanner', desc: 'Vulnerability scanning for containers and repos', icon: ICONS.shieldAlert, url: '/admin', group: 'Infrastructure', keywords: 'trivy vulnerability scan container image security' },
        { title: 'Autopilot', desc: 'Automated cluster health with dead server cleanup', icon: ICONS.gear, url: '/admin', group: 'Infrastructure', keywords: 'autopilot cluster health quorum cleanup' },
        { title: 'SafeGate Scanner', desc: 'File scanning pipeline with MIME, SVG, macro detection', icon: ICONS.shield, url: '/object-storage', group: 'Infrastructure', keywords: 'safegate scanner mime svg macro antivirus archive' },

        // Advanced
        { title: 'Vector Search', desc: 'Semantic search with vector embeddings and KNN', icon: ICONS.search, url: '/admin', group: 'Advanced', keywords: 'vector search embedding knn semantic ai' },
        { title: 'K8s Extensions', desc: 'Admission webhooks, CRD registry, scheduler', icon: ICONS.server, url: '/admin', group: 'Advanced', keywords: 'kubernetes k8s admission webhook crd scheduler' },
        { title: 'ML Pipeline', desc: 'Model deployment, feature extraction, training', icon: ICONS.gear, url: '/admin', group: 'Advanced', keywords: 'ml machine learning pipeline model training deploy' },
        { title: 'Prometheus Metrics', desc: 'Per-module metric collectors with /metrics endpoint', icon: ICONS.chart, url: '/admin', group: 'Advanced', keywords: 'prometheus metrics observability monitor' },

        // Pages
        { title: 'Admin Dashboard', desc: 'Full platform administration and management', icon: ICONS.home, url: '/admin', group: 'Pages', keywords: 'admin dashboard management' },
        { title: 'Sign Up', desc: 'Create a free account', icon: ICONS.userPlus, url: '/signup', group: 'Pages', keywords: 'signup register account create' },
        { title: 'Sign In', desc: 'Sign in to your account', icon: ICONS.lock, url: '/login', group: 'Pages', keywords: 'login signin sign-in auth' },
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
                closeSearch();
                requireAuth(href);
            }
            return;
        }
    });

    // Event: result click (use requireAuth for proper redirect)
    if (searchResults) {
        searchResults.addEventListener('click', function(e) {
            var result = e.target.closest('.search-result');
            if (result) {
                e.preventDefault();
                var href = result.getAttribute('href');
                closeSearch();
                requireAuth(href);
            }
        });

        // Event: result hover
        searchResults.addEventListener('mouseover', function(e) {
            var result = e.target.closest('.search-result');
            if (result) {
                var idx = parseInt(result.getAttribute('data-index'), 10);
                if (!isNaN(idx)) setActiveResult(idx);
            }
        });
    }

    // ============================================================
    // DATA FLOW CANVAS ANIMATION
    // ============================================================
    var dfCanvas = document.getElementById('dataflowCanvas');
    if (dfCanvas) {
        var dfCtx = dfCanvas.getContext('2d');
        var dfParticles = [];
        var dfPaths = [
            // Source -> Transform
            [{x: 0.08, y: 0.2}, {x: 0.15, y: 0.35}, {x: 0.28, y: 0.55}],
            // Transform -> Process
            [{x: 0.28, y: 0.55}, {x: 0.38, y: 0.35}, {x: 0.50, y: 0.25}],
            // Process -> Store
            [{x: 0.50, y: 0.25}, {x: 0.58, y: 0.45}, {x: 0.70, y: 0.6}],
            // Store -> Serve
            [{x: 0.70, y: 0.6}, {x: 0.78, y: 0.35}, {x: 0.90, y: 0.3}],
        ];

        function resizeDfCanvas() {
            dfCanvas.width = dfCanvas.offsetWidth;
            dfCanvas.height = dfCanvas.offsetHeight;
        }
        resizeDfCanvas();
        window.addEventListener('resize', resizeDfCanvas);

        // Spawn particles along paths
        function spawnDfParticle() {
            var pathIdx = Math.floor(Math.random() * dfPaths.length);
            var path = dfPaths[pathIdx];
            dfParticles.push({
                path: path,
                t: 0,
                speed: 0.005 + Math.random() * 0.008,
                size: Math.random() * 2.5 + 1.5,
                hue: pathIdx === 0 ? 160 : pathIdx === 1 ? 190 : pathIdx === 2 ? 260 : 160
            });
        }

        function drawDf() {
            dfCtx.clearRect(0, 0, dfCanvas.width, dfCanvas.height);

            // Draw paths
            for (var p = 0; p < dfPaths.length; p++) {
                var path = dfPaths[p];
                dfCtx.beginPath();
                dfCtx.moveTo(path[0].x * dfCanvas.width, path[0].y * dfCanvas.height);
                for (var pt = 1; pt < path.length; pt++) {
                    var prev = path[pt - 1];
                    var curr = path[pt];
                    var cpx = (prev.x + curr.x) / 2 * dfCanvas.width;
                    var cpy = (prev.y + curr.y) / 2 * dfCanvas.height;
                    dfCtx.quadraticCurveTo(prev.x * dfCanvas.width, prev.y * dfCanvas.height, cpx, cpy);
                }
                var last = path[path.length - 1];
                dfCtx.lineTo(last.x * dfCanvas.width, last.y * dfCanvas.height);
                dfCtx.strokeStyle = 'rgba(52, 211, 153, 0.06)';
                dfCtx.lineWidth = 2;
                dfCtx.stroke();
            }

            // Draw & update particles
            for (var i = dfParticles.length - 1; i >= 0; i--) {
                var particle = dfParticles[i];
                particle.t += particle.speed;
                if (particle.t >= 1) {
                    dfParticles.splice(i, 1);
                    continue;
                }

                var path = particle.path;
                var segCount = path.length - 1;
                var segT = particle.t * segCount;
                var segIdx = Math.min(Math.floor(segT), segCount - 1);
                var localT = segT - segIdx;
                var from = path[segIdx];
                var to = path[segIdx + 1];
                var px = (from.x + (to.x - from.x) * localT) * dfCanvas.width;
                var py = (from.y + (to.y - from.y) * localT) * dfCanvas.height;

                // Glow
                dfCtx.beginPath();
                dfCtx.arc(px, py, particle.size + 4, 0, Math.PI * 2);
                dfCtx.fillStyle = 'hsla(' + particle.hue + ', 80%, 55%, 0.15)';
                dfCtx.fill();
                // Core
                dfCtx.beginPath();
                dfCtx.arc(px, py, particle.size, 0, Math.PI * 2);
                dfCtx.fillStyle = 'hsla(' + particle.hue + ', 80%, 55%, 0.8)';
                dfCtx.fill();
            }

            requestAnimationFrame(drawDf);
        }

        // Start when visible
        var dfObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                drawDf();
                setInterval(spawnDfParticle, 200);
                dfObserver.unobserve(dfCanvas);
            }
        }, { threshold: 0.2 });
        dfObserver.observe(dfCanvas);
    }

    // ============================================================
    // API LIFECYCLE ANIMATION
    // ============================================================
    var apiStage = document.getElementById('apiLifecycleStage');
    var apiTrack = document.getElementById('apiTrack');
    var liveReqCode = document.getElementById('liveReqCode');
    var liveResCode = document.getElementById('liveResCode');
    var liveTiming = document.getElementById('liveTiming');

    var apiEndpoints = [
        { method: 'GET', url: '/api/v1/users', status: 200, latency: 23, resBody: '{\n  "data": [\n    { "id": 1, "name": "John Doe" },\n    { "id": 2, "name": "Jane Smith" }\n  ],\n  "total": 2\n}' },
        { method: 'POST', url: '/api/v1/orders', status: 201, latency: 45, resBody: '{\n  "id": "ord_9f3a2b",\n  "status": "created",\n  "total": 149.99\n}' },
        { method: 'GET', url: '/api/v1/products?page=1', status: 200, latency: 18, resBody: '{\n  "data": [\n    { "id": 1, "name": "Widget A", "price": 29.99 },\n    { "id": 2, "name": "Widget B", "price": 49.99 }\n  ],\n  "page": 1\n}' },
        { method: 'PUT', url: '/api/v1/users/1', status: 200, latency: 31, resBody: '{\n  "id": 1,\n  "name": "John Updated",\n  "updated_at": "2026-05-28T10:30:00Z"\n}' },
        { method: 'DELETE', url: '/api/v1/sessions/abc', status: 204, latency: 12, resBody: '' },
        { method: 'GET', url: '/api/v1/analytics/summary', status: 200, latency: 67, resBody: '{\n  "requests_today": 12847,\n  "avg_latency_ms": 23,\n  "error_rate": 0.02\n}' },
    ];

    var apiNodeIds = ['nodeClient', 'nodeGateway', 'nodeBuilder', 'nodeDB', 'nodeResponse'];
    var apiCycleIdx = 0;
    var apiCycleTimer = null;

    function spawnPacket(reverse) {
        if (!apiTrack) return;
        var dot = document.createElement('div');
        dot.className = 'api-lifecycle__packet' + (reverse ? ' api-lifecycle__packet--response' : '');
        apiTrack.appendChild(dot);
        setTimeout(function() { dot.remove(); }, 2100);
    }

    function activateNode(idx) {
        var nodes = apiStage ? apiStage.querySelectorAll('.api-lifecycle__node') : [];
        nodes.forEach(function(n, i) {
            n.classList.toggle('active', i <= idx);
        });
    }

    function runApiCycle() {
        var ep = apiEndpoints[apiCycleIdx % apiEndpoints.length];
        apiCycleIdx++;

        // Reset nodes
        activateNode(-1);
        if (liveReqCode) liveReqCode.innerHTML = '<span class="cursor-blink">|</span>';
        if (liveResCode) liveResCode.innerHTML = '';
        if (liveTiming) liveTiming.textContent = '';

        // Animate request flowing through nodes
        var steps = [
            { delay: 200, node: 0, text: '<span class="req-method">' + ep.method + '</span> <span class="req-url">' + ep.url + '</span>\n<span class="req-key">Authorization:</span> <span class="req-val">Bearer eyJhbG...</span>\n<span class="req-key">Content-Type:</span> <span class="req-val">application/json</span>' },
            { delay: 600, node: 1, text: null },
            { delay: 1000, node: 2, text: null },
            { delay: 1400, node: 3, text: null },
        ];

        steps.forEach(function(step) {
            setTimeout(function() {
                activateNode(step.node);
                spawnPacket(false);
                if (step.text && liveReqCode) {
                    liveReqCode.innerHTML = step.text;
                }
            }, step.delay);
        });

        // Response flows back
        setTimeout(function() {
            activateNode(4);
            spawnPacket(true);
            if (liveResCode) {
                if (ep.status === 204) {
                    liveResCode.innerHTML = '<span class="res-status">' + ep.status + '</span> No Content';
                } else {
                    // Syntax highlight the JSON
                    var highlighted = ep.resBody
                        .replace(/"([^"]+)":/g, '<span class="res-key">"$1"</span>:')
                        .replace(/: "([^"]+)"/g, ': <span class="res-str">"$1"</span>')
                        .replace(/: (\d+\.?\d*)/g, ': <span class="res-num">$1</span>');
                    liveResCode.innerHTML = '<span class="res-status">' + ep.status + ' ' + (ep.status === 200 ? 'OK' : ep.status === 201 ? 'Created' : 'OK') + '</span>\n' + highlighted;
                }
            }
            if (liveTiming) {
                liveTiming.textContent = ep.latency + 'ms';
            }
        }, 1800);
    }

    if (apiStage) {
        var apiObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                runApiCycle();
                apiCycleTimer = setInterval(runApiCycle, 4000);
                apiObserver.unobserve(apiStage);
            } else {
                clearInterval(apiCycleTimer);
            }
        }, { threshold: 0.3 });
        apiObserver.observe(apiStage);
    }

    // ============================================================
    // INTERACTIVE CLI TERMINAL
    // ============================================================
    var termBody = document.getElementById('terminalBody');
    var termInput = document.getElementById('terminalInput');
    var termAutocomplete = document.getElementById('terminalAutocomplete');

    var cliCommands = [
        { cmd: 'help', desc: 'Show all available commands' },
        { cmd: 'api create', desc: 'Create a new REST API (api create <name> --db <type> --table <name>)' },
        { cmd: 'api list', desc: 'List all created APIs' },
        { cmd: 'api delete', desc: 'Delete an API (api delete <name>)' },
        { cmd: 'api describe', desc: 'Show API details (api describe <name>)' },
        { cmd: 'db connect', desc: 'Connect to a database (db connect <type> --host <h> --port <p>)' },
        { cmd: 'db list', desc: 'List connected databases' },
        { cmd: 'db query', desc: 'Execute SQL query (db query --sql "SELECT ...")' },
        { cmd: 'storage bucket create', desc: 'Create a storage bucket (storage bucket create <name>)' },
        { cmd: 'storage bucket list', desc: 'List all storage buckets' },
        { cmd: 'storage object upload', desc: 'Upload an object (storage object upload <bucket> <file>)' },
        { cmd: 'workflow run', desc: 'Run an ETL workflow (workflow run <name> --input <file>)' },
        { cmd: 'workflow list', desc: 'List available workflows' },
        { cmd: 'workflow status', desc: 'Check workflow status (workflow status <id>)' },
        { cmd: 'cdc pipeline create', desc: 'Create a CDC pipeline (cdc pipeline create --source <s> --sink <s>)' },
        { cmd: 'cdc pipeline list', desc: 'List CDC pipelines' },
        { cmd: 'policy apply', desc: 'Apply a policy (policy apply <file.yaml>)' },
        { cmd: 'policy list', desc: 'List active policies' },
        { cmd: 'user create', desc: 'Create a user (user create --name <n> --email <e>)' },
        { cmd: 'user list', desc: 'List users' },
        { cmd: 'role assign', desc: 'Assign role to user (role assign <user> <role>)' },
        { cmd: 'metrics show', desc: 'Show platform metrics' },
        { cmd: 'health check', desc: 'Run health check on all services' },
        { cmd: 'config get', desc: 'Get configuration value (config get <key>)' },
        { cmd: 'config set', desc: 'Set configuration value (config set <key> <value>)' },
        { cmd: 'version', desc: 'Show axiomnizamctl version' },
        { cmd: 'clear', desc: 'Clear terminal output' },
    ];

    var termHistory = [];
    var termHistoryIdx = -1;
    var autocompleteIdx = -1;

    function termPrint(html, cls) {
        var div = document.createElement('div');
        div.className = 'terminal__line' + (cls ? ' ' + cls : '');
        div.innerHTML = html;
        if (termBody) termBody.appendChild(div);
        if (termBody) termBody.scrollTop = termBody.scrollHeight;
    }

    function executeCommand(input) {
        var raw = input.trim();
        if (!raw) return;
        termHistory.unshift(raw);
        if (termHistory.length > 50) termHistory.pop();
        termHistoryIdx = -1;

        termPrint('<span class="terminal__prompt">$</span> ' + escapeHTML(raw));

        var parts = raw.toLowerCase().split(/\s+/);
        var cmd = parts[0];
        var sub = parts[1] || '';
        var fullCmd = cmd + (sub ? ' ' + sub : '');

        if (cmd === 'clear') {
            if (termBody) termBody.innerHTML = '';
            return;
        }

        if (cmd === 'help') {
            termPrint('<span class="terminal__output-accent">Available commands:</span>');
            for (var i = 0; i < cliCommands.length; i++) {
                termPrint('  <span class="terminal__output-accent">' + cliCommands[i].cmd.padEnd(24) + '</span> <span class="terminal__output-dim">' + cliCommands[i].desc + '</span>');
            }
            return;
        }

        if (cmd === 'version') {
            termPrint('<span class="terminal__output-accent">axiomnizamctl</span> <span class="terminal__output-dim">v0.2.0</span>');
            termPrint('<span class="terminal__output-accent">AxiomNizam Platform</span> <span class="terminal__output-dim">v0.5.0</span>');
            termPrint('<span class="terminal__output-dim">Built: 2026-05-28 | Commit: baa1fff | Go: 1.24</span>');
            return;
        }

        if (cmd === 'api') {
            if (sub === 'create') {
                var name = parts[2] || 'my-api';
                var db = parts.indexOf('--db') > -1 ? parts[parts.indexOf('--db') + 1] : 'postgres';
                var table = parts.indexOf('--table') > -1 ? parts[parts.indexOf('--table') + 1] : 'default';
                termPrint('<span class="terminal__output-success">api "' + name + '" created</span>');
                termPrint('<span class="terminal__output-dim">  Database: ' + db + ' | Table: ' + table + ' | Endpoints: 5 CRUD</span>');
                termPrint('<span class="terminal__output-dim">  GET /api/v1/' + name + '  |  POST /api/v1/' + name + '  |  GET /api/v1/' + name + '/:id</span>');
                termPrint('<span class="terminal__output-dim">  PUT /api/v1/' + name + '/:id  |  DELETE /api/v1/' + name + '/:id</span>');
            } else if (sub === 'list') {
                termPrint('<span class="terminal__output-table">NAME          ENDPOINTS  DATABASE    STATUS</span>');
                termPrint('<span class="terminal__output-table">user-api      5          postgres    active</span>');
                termPrint('<span class="terminal__output-table">product-api   5          mysql       active</span>');
                termPrint('<span class="terminal__output-table">order-api     8          mongodb     active</span>');
            } else if (sub === 'delete') {
                var delName = parts[2] || '';
                termPrint('<span class="terminal__output-success">api "' + delName + '" deleted</span>');
            } else if (sub === 'describe') {
                var descName = parts[2] || '';
                termPrint('<span class="terminal__output-accent">API: ' + descName + '</span>');
                termPrint('<span class="terminal__output-dim">  Status: active | Created: 2026-05-28</span>');
                termPrint('<span class="terminal__output-dim">  Database: postgres | Table: users</span>');
                termPrint('<span class="terminal__output-dim">  Endpoints: 5 | Rate Limit: 1000/min</span>');
            } else {
                termPrint('<span class="terminal__output-error">Unknown api subcommand: ' + escapeHTML(sub) + '</span>');
                termPrint('<span class="terminal__output-dim">Try: api create, api list, api delete, api describe</span>');
            }
            return;
        }

        if (cmd === 'db') {
            if (sub === 'list') {
                termPrint('<span class="terminal__output-table">NAME          TYPE        HOST           STATUS</span>');
                termPrint('<span class="terminal__output-table">primary       postgres    localhost:5432  connected</span>');
                termPrint('<span class="terminal__output-table">analytics     mysql       localhost:3306  connected</span>');
                termPrint('<span class="terminal__output-table">nosql         mongodb     localhost:27017 connected</span>');
            } else if (sub === 'connect') {
                termPrint('<span class="terminal__output-success">Connected to database successfully</span>');
            } else if (sub === 'query') {
                termPrint('<span class="terminal__output-table">id  | name       | email</span>');
                termPrint('<span class="terminal__output-table">----|------------|------------------</span>');
                termPrint('<span class="terminal__output-table">1   | John Doe   | john@example.com</span>');
                termPrint('<span class="terminal__output-table">2   | Jane Smith | jane@example.com</span>');
                termPrint('<span class="terminal__output-dim">2 rows returned in 12ms</span>');
            } else {
                termPrint('<span class="terminal__output-error">Unknown db subcommand: ' + escapeHTML(sub) + '</span>');
            }
            return;
        }

        if (cmd === 'storage') {
            if (sub === 'bucket' && parts[2] === 'list') {
                termPrint('<span class="terminal__output-table">NAME          OBJECTS  SIZE      ENCRYPTED</span>');
                termPrint('<span class="terminal__output-table">uploads       1,247    2.3 GB    yes</span>');
                termPrint('<span class="terminal__output-table">backups       89       456 MB    yes</span>');
                termPrint('<span class="terminal__output-table">media         3,456    12.1 GB   no</span>');
            } else if (sub === 'bucket' && parts[2] === 'create') {
                var bname = parts[3] || 'new-bucket';
                termPrint('<span class="terminal__output-success">bucket "' + bname + '" created</span>');
            } else {
                termPrint('<span class="terminal__output-dim">Usage: storage bucket create|list [name]</span>');
            }
            return;
        }

        if (cmd === 'workflow') {
            if (sub === 'run') {
                var wfName = parts[2] || 'pipeline';
                termPrint('<span class="terminal__output-success">workflow "' + wfName + '" started</span>');
                termPrint('<span class="terminal__output-dim">  ID: wf-' + Math.random().toString(36).substr(2, 8) + '</span>');
                termPrint('<span class="terminal__output-dim">  Records: 1,247 processed | Duration: 3.2s</span>');
            } else if (sub === 'list') {
                termPrint('<span class="terminal__output-table">NAME            STATUS    LAST RUN</span>');
                termPrint('<span class="terminal__output-table">etl-pipeline    running   2 min ago</span>');
                termPrint('<span class="terminal__output-table">data-sync       idle      1 hour ago</span>');
            } else if (sub === 'status') {
                termPrint('<span class="terminal__output-accent">Workflow Status: running</span>');
                termPrint('<span class="terminal__output-dim">  Progress: 78% | Records: 973/1,247</span>');
            } else {
                termPrint('<span class="terminal__output-error">Unknown workflow subcommand</span>');
            }
            return;
        }

        if (cmd === 'cdc') {
            termPrint('<span class="terminal__output-success">CDC pipeline operation completed</span>');
            return;
        }

        if (cmd === 'policy') {
            if (sub === 'list') {
                termPrint('<span class="terminal__output-table">NAME            TYPE        STATUS</span>');
                termPrint('<span class="terminal__output-table">compliance      access      active</span>');
                termPrint('<span class="terminal__output-table">rate-limit      quota       active</span>');
                termPrint('<span class="terminal__output-table">data-retention  lifecycle   active</span>');
            } else if (sub === 'apply') {
                termPrint('<span class="terminal__output-success">policy applied — 3 rules active</span>');
            } else {
                termPrint('<span class="terminal__output-error">Unknown policy subcommand</span>');
            }
            return;
        }

        if (cmd === 'user') {
            if (sub === 'list') {
                termPrint('<span class="terminal__output-table">NAME          EMAIL               ROLE</span>');
                termPrint('<span class="terminal__output-table">admin         admin@axiom.io      admin</span>');
                termPrint('<span class="terminal__output-table">developer     dev@axiom.io        manager</span>');
            } else {
                termPrint('<span class="terminal__output-success">User operation completed</span>');
            }
            return;
        }

        if (cmd === 'role') {
            termPrint('<span class="terminal__output-success">Role operation completed</span>');
            return;
        }

        if (cmd === 'metrics') {
            termPrint('<span class="terminal__output-accent">Platform Metrics:</span>');
            termPrint('<span class="terminal__output-dim">  API Requests:  12,847/min</span>');
            termPrint('<span class="terminal__output-dim">  Avg Latency:   23ms</span>');
            termPrint('<span class="terminal__output-dim">  Error Rate:    0.02%</span>');
            termPrint('<span class="terminal__output-dim">  Uptime:        99.97%</span>');
            return;
        }

        if (cmd === 'health') {
            termPrint('<span class="terminal__output-accent">Health Check:</span>');
            termPrint('  <span class="terminal__output-success">OK</span>  API Server');
            termPrint('  <span class="terminal__output-success">OK</span>  Database');
            termPrint('  <span class="terminal__output-success">OK</span>  Storage');
            termPrint('  <span class="terminal__output-success">OK</span>  Message Queue');
            termPrint('  <span class="terminal__output-success">OK</span>  Cache');
            termPrint('<span class="terminal__output-dim">All services healthy</span>');
            return;
        }

        if (cmd === 'config') {
            termPrint('<span class="terminal__output-success">Config operation completed</span>');
            return;
        }

        termPrint('<span class="terminal__output-error">Command not found: ' + escapeHTML(cmd) + '</span>');
        termPrint('<span class="terminal__output-dim">Type "help" to see available commands</span>');
    }

    function updateAutocomplete(value) {
        if (!termAutocomplete) return;
        if (!value || value.length < 1) {
            termAutocomplete.classList.remove('visible');
            termAutocomplete.innerHTML = '';
            autocompleteIdx = -1;
            return;
        }

        var matches = [];
        var lower = value.toLowerCase();
        for (var i = 0; i < cliCommands.length; i++) {
            if (cliCommands[i].cmd.toLowerCase().indexOf(lower) === 0 || cliCommands[i].cmd.toLowerCase().indexOf(lower) > -1) {
                matches.push(cliCommands[i]);
            }
        }

        if (matches.length === 0) {
            termAutocomplete.classList.remove('visible');
            termAutocomplete.innerHTML = '';
            autocompleteIdx = -1;
            return;
        }

        var html = '<div class="terminal__autocomplete-hint"><kbd>Tab</kbd> to complete &middot; <kbd>&uarr;&darr;</kbd> to navigate &middot; <kbd>Enter</kbd> to select</div>';
        for (var j = 0; j < matches.length; j++) {
            html += '<div class="terminal__autocomplete-item' + (j === 0 ? ' active' : '') + '" data-cmd="' + matches[j].cmd + '" data-index="' + j + '">'
                + '<span class="terminal__autocomplete-cmd">' + matches[j].cmd + '</span>'
                + '<span class="terminal__autocomplete-desc">' + matches[j].desc + '</span>'
                + '</div>';
        }
        termAutocomplete.innerHTML = html;
        termAutocomplete.classList.add('visible');
        autocompleteIdx = 0;
    }

    function selectAutocompleteItem(idx) {
        var items = termAutocomplete ? termAutocomplete.querySelectorAll('.terminal__autocomplete-item') : [];
        if (items.length === 0) return;
        if (idx < 0) idx = items.length - 1;
        if (idx >= items.length) idx = 0;
        for (var i = 0; i < items.length; i++) {
            items[i].classList.toggle('active', i === idx);
        }
        autocompleteIdx = idx;
    }

    if (termInput) {
        termInput.addEventListener('input', function() {
            updateAutocomplete(termInput.value.trim());
        });

        termInput.addEventListener('keydown', function(e) {
            var acVisible = termAutocomplete && termAutocomplete.classList.contains('visible');

            // Tab completion
            if (e.key === 'Tab') {
                e.preventDefault();
                if (acVisible) {
                    var activeItem = termAutocomplete.querySelector('.terminal__autocomplete-item.active');
                    if (activeItem) {
                        termInput.value = activeItem.getAttribute('data-cmd') + ' ';
                        termAutocomplete.classList.remove('visible');
                        termInput.focus();
                    }
                }
                return;
            }

            // Arrow keys
            if (e.key === 'ArrowDown') {
                e.preventDefault();
                if (acVisible) {
                    selectAutocompleteItem(autocompleteIdx + 1);
                } else if (termHistory.length > 0) {
                    termHistoryIdx = Math.max(0, termHistoryIdx - 1);
                    termInput.value = termHistory[termHistoryIdx] || '';
                }
                return;
            }

            if (e.key === 'ArrowUp') {
                e.preventDefault();
                if (acVisible) {
                    selectAutocompleteItem(autocompleteIdx - 1);
                } else if (termHistory.length > 0) {
                    termHistoryIdx = Math.min(termHistory.length - 1, termHistoryIdx + 1);
                    termInput.value = termHistory[termHistoryIdx] || '';
                }
                return;
            }

            // Enter
            if (e.key === 'Enter') {
                e.preventDefault();
                var cmd = termInput.value.trim();
                if (acVisible) {
                    var selected = termAutocomplete.querySelector('.terminal__autocomplete-item.active');
                    if (selected) {
                        cmd = selected.getAttribute('data-cmd');
                        termInput.value = cmd + ' ';
                        termAutocomplete.classList.remove('visible');
                        // Don't execute on autocomplete select, let user finish typing args
                        return;
                    }
                }
                termInput.value = '';
                termAutocomplete.classList.remove('visible');
                executeCommand(cmd);
                return;
            }

            // Escape
            if (e.key === 'Escape') {
                termAutocomplete.classList.remove('visible');
                return;
            }
        });

        // Click autocomplete item
        if (termAutocomplete) {
            termAutocomplete.addEventListener('click', function(e) {
                var item = e.target.closest('.terminal__autocomplete-item');
                if (item) {
                    termInput.value = item.getAttribute('data-cmd') + ' ';
                    termAutocomplete.classList.remove('visible');
                    termInput.focus();
                }
            });
        }
    }

    // ============================================================
    // API DEMO TABS
    // ============================================================
    var apiDemoTabs = document.querySelectorAll('.api-demo__tab');
    var apiDemoPanels = document.querySelectorAll('.api-demo__panel');

    apiDemoTabs.forEach(function(tab) {
        tab.addEventListener('click', function(e) {
            e.stopPropagation();
            var targetId = 'demo-' + tab.getAttribute('data-demo');
            apiDemoTabs.forEach(function(t) { t.classList.remove('active'); });
            apiDemoPanels.forEach(function(p) { p.classList.remove('active'); });
            tab.classList.add('active');
            var panel = document.getElementById(targetId);
            if (panel) panel.classList.add('active');
        });
    });

    // Auto-cycle API demo panels
    var demoPanelNames = ['request', 'response', 'endpoints'];
    var demoCycleIdx = 0;
    var demoCycleTimer = null;

    function cycleApiDemo() {
        demoCycleIdx = (demoCycleIdx + 1) % demoPanelNames.length;
        var targetId = 'demo-' + demoPanelNames[demoCycleIdx];
        apiDemoTabs.forEach(function(t) {
            t.classList.toggle('active', t.getAttribute('data-demo') === demoPanelNames[demoCycleIdx]);
        });
        apiDemoPanels.forEach(function(p) { p.classList.remove('active'); });
        var panel = document.getElementById(targetId);
        if (panel) panel.classList.add('active');
    }

    // Start auto-cycle when visible
    var apiDemo = document.getElementById('apiDemo');
    if (apiDemo) {
        var demoObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                demoCycleTimer = setInterval(cycleApiDemo, 3500);
            } else {
                clearInterval(demoCycleTimer);
            }
        }, { threshold: 0.5 });
        demoObserver.observe(apiDemo);

        // Stop auto-cycle on manual click
        apiDemoTabs.forEach(function(tab) {
            tab.addEventListener('click', function() {
                clearInterval(demoCycleTimer);
                demoCycleIdx = demoPanelNames.indexOf(tab.getAttribute('data-demo'));
            });
        });
    }

    // ============================================================
    // STORAGE DEMO BAR ANIMATION
    // ============================================================
    var storageBars = document.querySelectorAll('.storage-demo__bar-fill');
    if (storageBars.length > 0) {
        var storageObserver = new IntersectionObserver(function(entries) {
            entries.forEach(function(entry) {
                if (entry.isIntersecting) {
                    var bars = entry.target.querySelectorAll('.storage-demo__bar-fill');
                    bars.forEach(function(bar, i) {
                        var width = bar.getAttribute('data-width') || '50';
                        setTimeout(function() {
                            bar.style.width = width + '%';
                        }, i * 200);
                    });
                    storageObserver.unobserve(entry.target);
                }
            });
        }, { threshold: 0.3 });
        var storageDemo = document.getElementById('storageDemo');
        if (storageDemo) storageObserver.observe(storageDemo);
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

                // Draw particle with glow
                ctx.beginPath();
                ctx.arc(p.x, p.y, p.radius + 2, 0, Math.PI * 2);
                ctx.fillStyle = 'rgba(52, 211, 153, ' + (p.opacity * 0.15) + ')';
                ctx.fill();
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

    // ---- Enhanced Cursor System ----
    var cursorGlow = document.getElementById('cursorGlow');
    var cursorRing = document.getElementById('cursorRing');
    var cursorCrosshair = document.getElementById('cursorCrosshair');
    var cursorTrailCanvas = document.getElementById('cursorTrail');

    var cursorX = 0, cursorY = 0;
    var glowX = 0, glowY = 0;
    var ringX = 0, ringY = 0;
    var crossX = 0, crossY = 0;
    var trailParticles = [];
    var isMouseInDoc = false;

    document.addEventListener('mousemove', function(e) {
        cursorX = e.clientX;
        cursorY = e.clientY;
        isMouseInDoc = true;
        if (cursorGlow) cursorGlow.classList.add('active');
        if (cursorRing) cursorRing.classList.add('active');
        if (cursorCrosshair) cursorCrosshair.classList.add('active');

        // Add trail particle
        trailParticles.push({
            x: cursorX,
            y: cursorY,
            life: 1,
            vx: (Math.random() - 0.5) * 1.5,
            vy: (Math.random() - 0.5) * 1.5,
            size: Math.random() * 2 + 1
        });
        if (trailParticles.length > 50) trailParticles.shift();
    });

    document.addEventListener('mouseleave', function() {
        isMouseInDoc = false;
        if (cursorGlow) cursorGlow.classList.remove('active');
        if (cursorRing) cursorRing.classList.remove('active');
        if (cursorCrosshair) cursorCrosshair.classList.remove('active');
    });

    // Click ripple effect
    document.addEventListener('mousedown', function(e) {
        if (cursorRing) cursorRing.classList.add('clicking');
        // Create ripple
        var ripple = document.createElement('div');
        ripple.className = 'cursor-ripple';
        ripple.style.left = e.clientX + 'px';
        ripple.style.top = e.clientY + 'px';
        document.body.appendChild(ripple);
        setTimeout(function() { ripple.remove(); }, 600);
        // Burst particles
        for (var b = 0; b < 8; b++) {
            var angle = (Math.PI * 2 / 8) * b;
            trailParticles.push({
                x: e.clientX,
                y: e.clientY,
                vx: Math.cos(angle) * 3.5,
                vy: Math.sin(angle) * 3.5,
                life: 1,
                size: Math.random() * 2.5 + 1.5
            });
        }
    });
    document.addEventListener('mouseup', function() {
        if (cursorRing) cursorRing.classList.remove('clicking');
    });

    // Detect hoverable elements for ring expansion + particle burst
    var lastHoverTarget = null;
    document.addEventListener('mouseover', function(e) {
        var target = e.target;
        if (target.closest('a, button, [data-tilt], .search-trigger, .deep__tab, .bento__card, .arch__card, .search-result, .terminal__autocomplete-item, .api-demo__tab')) {
            if (cursorRing) cursorRing.classList.add('hovering');
            // Mini particle burst on hover enter
            var el = target.closest('a, button, [data-tilt], .search-trigger, .deep__tab, .bento__card, .arch__card, .search-result, .terminal__autocomplete-item, .api-demo__tab');
            if (el !== lastHoverTarget) {
                lastHoverTarget = el;
                for (var h = 0; h < 5; h++) {
                    trailParticles.push({
                        x: cursorX + (Math.random() - 0.5) * 20,
                        y: cursorY + (Math.random() - 0.5) * 20,
                        vx: (Math.random() - 0.5) * 2,
                        vy: (Math.random() - 0.5) * 2,
                        life: 0.8,
                        size: Math.random() * 2 + 1
                    });
                }
            }
        }
    });
    document.addEventListener('mouseout', function(e) {
        var target = e.target;
        if (target.closest('a, button, [data-tilt], .search-trigger, .deep__tab, .bento__card, .arch__card, .search-result, .terminal__autocomplete-item, .api-demo__tab')) {
            if (cursorRing) cursorRing.classList.remove('hovering');
            lastHoverTarget = null;
        }
    });

    // Trail canvas setup
    if (cursorTrailCanvas) {
        var trailCtx = cursorTrailCanvas.getContext('2d');
        function resizeTrailCanvas() {
            cursorTrailCanvas.width = window.innerWidth;
            cursorTrailCanvas.height = window.innerHeight;
        }
        resizeTrailCanvas();
        window.addEventListener('resize', resizeTrailCanvas);
    }

    function animateCursorSystem() {
        // Smooth follow for glow
        glowX += (cursorX - glowX) * 0.06;
        glowY += (cursorY - glowY) * 0.06;
        if (cursorGlow) {
            cursorGlow.style.left = glowX + 'px';
            cursorGlow.style.top = glowY + 'px';
        }

        // Ring follows with more delay
        ringX += (cursorX - ringX) * 0.15;
        ringY += (cursorY - ringY) * 0.15;
        if (cursorRing) {
            cursorRing.style.left = ringX + 'px';
            cursorRing.style.top = ringY + 'px';
        }

        // Crosshair follows with medium delay
        crossX += (cursorX - crossX) * 0.1;
        crossY += (cursorY - crossY) * 0.1;
        if (cursorCrosshair) {
            cursorCrosshair.style.left = crossX + 'px';
            cursorCrosshair.style.top = crossY + 'px';
        }

        // Draw trail particles with glow and flowing line
        if (cursorTrailCanvas && trailCtx) {
            trailCtx.clearRect(0, 0, cursorTrailCanvas.width, cursorTrailCanvas.height);

            // Draw flowing cursor trail line
            if (trailParticles.length > 3 && isMouseInDoc) {
                var recent = trailParticles.slice(-15);
                trailCtx.beginPath();
                trailCtx.moveTo(recent[0].x, recent[0].y);
                for (var t = 1; t < recent.length - 1; t++) {
                    var cpx = (recent[t].x + recent[t + 1].x) / 2;
                    var cpy = (recent[t].y + recent[t + 1].y) / 2;
                    trailCtx.quadraticCurveTo(recent[t].x, recent[t].y, cpx, cpy);
                }
                var lastP = recent[recent.length - 1];
                trailCtx.lineTo(lastP.x, lastP.y);
                trailCtx.strokeStyle = 'rgba(52, 211, 153, 0.12)';
                trailCtx.lineWidth = 2;
                trailCtx.stroke();
            }
            for (var i = trailParticles.length - 1; i >= 0; i--) {
                var p = trailParticles[i];
                p.x += p.vx;
                p.y += p.vy;
                p.vx *= 0.98;
                p.vy *= 0.98;
                p.life -= 0.02;
                if (p.life <= 0) {
                    trailParticles.splice(i, 1);
                    continue;
                }
                trailCtx.beginPath();
                trailCtx.arc(p.x, p.y, p.size * p.life, 0, Math.PI * 2);
                trailCtx.fillStyle = 'rgba(52, 211, 153, ' + (p.life * 0.4) + ')';
                trailCtx.fill();
            }

            // Draw connections between nearby trail particles
            for (var j = 0; j < trailParticles.length; j++) {
                for (var k = j + 1; k < trailParticles.length; k++) {
                    var a = trailParticles[j];
                    var b = trailParticles[k];
                    var dx = a.x - b.x;
                    var dy = a.y - b.y;
                    var dist = Math.sqrt(dx * dx + dy * dy);
                    if (dist < 40) {
                        var alpha = (1 - dist / 40) * a.life * b.life * 0.2;
                        trailCtx.beginPath();
                        trailCtx.moveTo(a.x, a.y);
                        trailCtx.lineTo(b.x, b.y);
                        trailCtx.strokeStyle = 'rgba(52, 211, 153, ' + alpha + ')';
                        trailCtx.lineWidth = 0.5;
                        trailCtx.stroke();
                    }
                }
            }
        }

        requestAnimationFrame(animateCursorSystem);
    }
    animateCursorSystem();

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
    var heroLogoBg = document.getElementById('heroLogoBg');
    var heroLogoSvg = document.getElementById('heroLogoSvg');

    if (heroSection) {
        heroSection.addEventListener('mousemove', function(e) {
            var rect = heroSection.getBoundingClientRect();
            var x = (e.clientX - rect.left) / rect.width - 0.5;
            var y = (e.clientY - rect.top) / rect.height - 0.5;

            // Orbs react to mouse
            heroOrbs.forEach(function(orb, i) {
                var depth = (i + 1) * 8;
                orb.style.transform = 'translate(' + (x * depth) + 'px, ' + (y * depth) + 'px)';
            });

            // Logo subtle mouse follow
            if (heroLogoBg) {
                heroLogoBg.style.transform = 'translate(calc(-50% + ' + (x * 15) + 'px), calc(-50% + ' + (y * 15) + 'px))';
            }
        });

        // Logo parallax on scroll — fades as you scroll
        var logoParallaxTicking = false;
        window.addEventListener('scroll', function() {
            if (!logoParallaxTicking) {
                requestAnimationFrame(function() {
                    var scrollY = window.pageYOffset;
                    var heroH = heroSection.offsetHeight;
                    var progress = Math.min(scrollY / heroH, 1);
                    if (heroLogoBg) {
                        heroLogoBg.style.opacity = Math.max(1 - progress * 1.5, 0);
                    }
                    logoParallaxTicking = false;
                });
                logoParallaxTicking = true;
            }
        });
    }

    // ============================================================
    // LEGO ASSEMBLY / DISASSEMBLY ANIMATION
    // ============================================================
    var logoPieces = document.querySelectorAll('.logo-piece');
    var legoState = 'assembled'; // assembled | disassembled
    var legoTimer = null;

    function legoDisassemble() {
        legoState = 'disassembled';
        logoPieces.forEach(function(piece, i) {
            var fromX = parseFloat(piece.getAttribute('data-from-x')) || 0;
            var fromY = parseFloat(piece.getAttribute('data-from-y')) || 0;
            var fromR = parseFloat(piece.getAttribute('data-from-r')) || 0;
            var fromS = parseFloat(piece.getAttribute('data-from-s')) || 0.5;

            // Stagger each piece
            setTimeout(function() {
                piece.style.transition = 'transform 0.8s cubic-bezier(0.68, -0.55, 0.27, 1.55), opacity 0.6s ease';
                piece.style.transform = 'translate(' + fromX + 'px, ' + fromY + 'px) rotate(' + fromR + 'deg) scale(' + fromS + ')';
                piece.style.opacity = '0.5';
            }, i * 80);
        });

        // After disassembled, wait then reassemble
        legoTimer = setTimeout(legoAssemble, 2500);
    }

    function legoAssemble() {
        legoState = 'assembled';
        logoPieces.forEach(function(piece, i) {
            // Stagger reassembly — pieces fly in
            setTimeout(function() {
                piece.style.transition = 'transform 1s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.8s ease';
                piece.style.transform = 'translate(0, 0) rotate(0deg) scale(1)';
                piece.style.opacity = '1';
            }, i * 120);
        });

        // After assembled, wait then disassemble again
        legoTimer = setTimeout(legoDisassemble, 5000);
    }

    // Start the cycle when hero is visible
    if (logoPieces.length > 0) {
        var legoObserver = new IntersectionObserver(function(entries) {
            if (entries[0].isIntersecting) {
                // Initial assembly animation — pieces fly in from scattered positions
                logoPieces.forEach(function(piece) {
                    var fromX = parseFloat(piece.getAttribute('data-from-x')) || 0;
                    var fromY = parseFloat(piece.getAttribute('data-from-y')) || 0;
                    var fromR = parseFloat(piece.getAttribute('data-from-r')) || 0;
                    var fromS = parseFloat(piece.getAttribute('data-from-s')) || 0.5;
                    piece.style.transform = 'translate(' + fromX + 'px, ' + fromY + 'px) rotate(' + fromR + 'deg) scale(' + fromS + ')';
                    piece.style.opacity = '0';
                });

                // Animate in after a short delay
                setTimeout(legoAssemble, 800);
                legoObserver.unobserve(heroSection);
            }
        }, { threshold: 0.3 });
        legoObserver.observe(heroSection);
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
