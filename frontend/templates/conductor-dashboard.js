(function() {
    'use strict';

    var API = window.resolveBackendURL() + '/api/v1/conductor';
    var wsURL = '';
    var streamWS = null;
    var streamConnected = false;
    var cachedProducers = [];
    var cachedConsumers = [];

    function authHeaders() {
        var token = '';
        try { token = document.cookie.split(';').map(function(c){return c.trim();}).find(function(c){return c.startsWith('authToken=');}) || ''; } catch(e){}
        if (token) token = token.split('=')[1] || '';
        var headers = { 'Content-Type': 'application/json' };
        if (token) headers['Authorization'] = 'Bearer ' + token;
        return headers;
    }

    function apiGet(path) {
        return fetch(API + path, { headers: authHeaders() }).then(function(r) { return r.json(); });
    }
    function apiPost(path, body) {
        return fetch(API + path, { method: 'POST', headers: authHeaders(), body: JSON.stringify(body) }).then(function(r) { return r.json(); });
    }
    function apiPatch(path, body) {
        return fetch(API + path, { method: 'PATCH', headers: authHeaders(), body: JSON.stringify(body) }).then(function(r) { return r.json(); });
    }
    function apiDelete(path) {
        return fetch(API + path, { method: 'DELETE', headers: authHeaders() }).then(function(r) { return r.json(); });
    }

    function statusBadge(status) {
        return '<span class="status-badge status-' + (status || 'unknown') + '">' + (status || 'unknown') + '</span>';
    }
    function backendBadge(backend) {
        return '<span class="backend-badge backend-' + (backend || 'memory') + '">' + (backend || 'memory') + '</span>';
    }
    function formatTime(ts) {
        if (!ts) return '-';
        var d = new Date(ts);
        if (isNaN(d.getTime())) return '-';
        return d.toLocaleString();
    }
    function jsonPreview(obj) {
        if (!obj) return '-';
        var s = JSON.stringify(obj);
        if (s.length > 80) s = s.substring(0, 80) + '...';
        return '<span class="json-preview" title="' + s.replace(/"/g, '&quot;') + '">' + s + '</span>';
    }

    // ---- Stats ----
    function loadStats() {
        apiGet('/stats').then(function(data) {
            document.getElementById('statProducers').textContent = data.producers || 0;
            document.getElementById('statConsumers').textContent = data.consumers || 0;
            document.getElementById('statSent').textContent = data.totalSent || 0;
            document.getElementById('statReceived').textContent = data.totalReceived || 0;
            document.getElementById('statAcked').textContent = data.totalAcked || 0;
            document.getElementById('statFailed').textContent = data.totalFailed || 0;
            document.getElementById('statDLQ').textContent = data.dlqSize || 0;
        }).catch(function() {});
    }

    // ---- Producers ----
    function loadProducers() {
        apiGet('/producers').then(function(data) {
            var prods = data.producers || [];
            cachedProducers = prods;
            var tbody = document.getElementById('producersBody');
            if (!prods.length) {
                tbody.innerHTML = '<tr><td colspan="8" class="empty-row">No producers configured yet</td></tr>';
                return;
            }
            tbody.innerHTML = prods.map(function(p) {
                var target = p.backend === 'kafka' ? (p.topic || '-') : (p.exchange || '-');
                return '<tr class="hoverable-row" data-type="producer" data-id="' + p.id + '">' +
                    '<td><strong>' + p.name + '</strong><br><small style="color:var(--text-muted)">' + p.id + '</small></td>' +
                    '<td>' + backendBadge(p.backend) + '</td>' +
                    '<td>' + target + '</td>' +
                    '<td>' + (p.routingKey || '-') + '</td>' +
                    '<td>' + statusBadge(p.status) + '</td>' +
                    '<td>' + (p.messagesSent || 0) + '</td>' +
                    '<td>' + formatTime(p.lastSentAt) + '</td>' +
                    '<td>' + producerActions(p) + '</td>' +
                    '</tr>';
            }).join('');
            updateProducerSelect(prods);
            bindHoverMetrics();
        }).catch(function() {});
    }

    function producerActions(p) {
        var html = '<button class="action-btn" onclick="window._conductorEditProducer(\'' + p.id + '\')" title="Edit">✏️</button>';
        if (p.status === 'active') {
            html += '<button class="action-btn" onclick="window._conductorPauseProducer(\'' + p.id + '\')">Pause</button>';
        } else if (p.status === 'paused') {
            html += '<button class="action-btn" onclick="window._conductorResumeProducer(\'' + p.id + '\')">Resume</button>';
        }
        html += '<button class="action-btn danger" onclick="window._conductorDeleteProducer(\'' + p.id + '\')">Delete</button>';
        return html;
    }

    function updateProducerSelect(prods) {
        var sel = document.getElementById('publishProducer');
        if (!sel) return;
        sel.innerHTML = prods.filter(function(p) { return p.status === 'active'; }).map(function(p) {
            return '<option value="' + p.id + '">' + p.name + ' (' + p.backend + ')</option>';
        }).join('');
    }

    window._conductorPauseProducer = function(id) {
        apiPost('/producers/' + id + '/pause', {}).then(function() { loadProducers(); loadStats(); });
    };
    window._conductorResumeProducer = function(id) {
        apiPost('/producers/' + id + '/resume', {}).then(function() { loadProducers(); loadStats(); });
    };
    window._conductorDeleteProducer = function(id) {
        if (!confirm('Delete this producer?')) return;
        apiDelete('/producers/' + id).then(function() { loadProducers(); loadStats(); });
    };

    window._conductorEditProducer = function(id) {
        var p = cachedProducers.find(function(x) { return x.id === id; });
        if (!p) return;
        document.getElementById('editProdId').value = p.id;
        document.getElementById('editProdName').value = p.name || '';
        document.getElementById('editProdExchange').value = (p.backend === 'kafka' ? p.topic : p.exchange) || '';
        document.getElementById('editProdRoutingKey').value = p.routingKey || '';
        document.getElementById('editProdContentType').value = p.contentType || 'application/json';
        document.getElementById('editProdPersistent').checked = p.config && p.config.persistent;
        document.getElementById('editProducerModal').style.display = 'flex';
    };

    window.saveEditProducer = function() {
        var id = document.getElementById('editProdId').value;
        var p = cachedProducers.find(function(x) { return x.id === id; });
        var exchTopic = document.getElementById('editProdExchange').value;
        var req = {
            name: document.getElementById('editProdName').value,
            backend: p ? p.backend : 'rabbitmq',
            contentType: document.getElementById('editProdContentType').value,
            config: { persistent: document.getElementById('editProdPersistent').checked }
        };
        if (p && p.backend === 'kafka') {
            req.topic = exchTopic;
        } else {
            req.exchange = exchTopic;
        }
        req.routingKey = document.getElementById('editProdRoutingKey').value;
        apiPatch('/producers/' + id, req).then(function() {
            closeModal('editProducerModal');
            loadProducers();
            loadStats();
        }).catch(function(e) { alert('Error: ' + e); });
    };

    window.showCreateProducerModal = function() {
        document.getElementById('createProducerModal').style.display = 'flex';
    };

    window.createProducer = function() {
        var backend = document.getElementById('prodBackend').value;
        var req = {
            name: document.getElementById('prodName').value,
            backend: backend,
            contentType: document.getElementById('prodContentType').value || 'application/json',
            config: { persistent: document.getElementById('prodPersistent').checked }
        };
        var exchTopic = document.getElementById('prodExchange').value;
        if (backend === 'kafka') {
            req.topic = exchTopic;
        } else {
            req.exchange = exchTopic;
        }
        req.routingKey = document.getElementById('prodRoutingKey').value;
        apiPost('/producers', req).then(function() {
            closeModal('createProducerModal');
            loadProducers();
            loadStats();
        }).catch(function(e) { alert('Error: ' + e); });
    };

    // ---- Consumers ----
    function loadConsumers() {
        apiGet('/consumers').then(function(data) {
            var cons = data.consumers || [];
            cachedConsumers = cons;
            var tbody = document.getElementById('consumersBody');
            if (!cons.length) {
                tbody.innerHTML = '<tr><td colspan="9" class="empty-row">No consumers configured yet</td></tr>';
                return;
            }
            tbody.innerHTML = cons.map(function(c) {
                var target = c.backend === 'kafka' ? (c.topic || '-') : (c.queue || '-');
                return '<tr class="hoverable-row" data-type="consumer" data-id="' + c.id + '">' +
                    '<td><strong>' + c.name + '</strong><br><small style="color:var(--text-muted)">' + c.id + '</small></td>' +
                    '<td>' + backendBadge(c.backend) + '</td>' +
                    '<td>' + target + '</td>' +
                    '<td>' + (c.consumerGroup || '-') + '</td>' +
                    '<td>' + statusBadge(c.status) + '</td>' +
                    '<td>' + (c.messagesReceived || 0) + '</td>' +
                    '<td>' + (c.messagesAcked || 0) + '</td>' +
                    '<td>' + (c.messagesFailed || 0) + '</td>' +
                    '<td>' + consumerActions(c) + '</td>' +
                    '</tr>';
            }).join('');
            bindHoverMetrics();
        }).catch(function() {});
    }

    function consumerActions(c) {
        var html = '<button class="action-btn" onclick="window._conductorEditConsumer(\'' + c.id + '\')" title="Edit">✏️</button>';
        if (c.status === 'active') {
            html += '<button class="action-btn" onclick="window._conductorPauseConsumer(\'' + c.id + '\')">Pause</button>';
        } else if (c.status === 'paused') {
            html += '<button class="action-btn" onclick="window._conductorResumeConsumer(\'' + c.id + '\')">Resume</button>';
        }
        html += '<button class="action-btn danger" onclick="window._conductorDeleteConsumer(\'' + c.id + '\')">Delete</button>';
        return html;
    }

    window._conductorPauseConsumer = function(id) {
        apiPost('/consumers/' + id + '/pause', {}).then(function() { loadConsumers(); loadStats(); });
    };
    window._conductorResumeConsumer = function(id) {
        apiPost('/consumers/' + id + '/resume', {}).then(function() { loadConsumers(); loadStats(); });
    };
    window._conductorDeleteConsumer = function(id) {
        if (!confirm('Delete this consumer?')) return;
        apiDelete('/consumers/' + id).then(function() { loadConsumers(); loadStats(); });
    };

    window._conductorEditConsumer = function(id) {
        var c = cachedConsumers.find(function(x) { return x.id === id; });
        if (!c) return;
        document.getElementById('editConsId').value = c.id;
        document.getElementById('editConsName').value = c.name || '';
        document.getElementById('editConsQueue').value = (c.backend === 'kafka' ? c.topic : c.queue) || '';
        document.getElementById('editConsExchange').value = c.exchange || '';
        document.getElementById('editConsRoutingKey').value = c.routingKey || '';
        document.getElementById('editConsGroup').value = c.consumerGroup || '';
        document.getElementById('editConsPrefetch').value = (c.config && c.config.prefetchCount) || 10;
        document.getElementById('editConsMaxRetries').value = (c.config && c.config.maxRetries) || 3;
        document.getElementById('editConsDLQ').checked = c.config && c.config.dlqEnabled;
        document.getElementById('editConsumerModal').style.display = 'flex';
    };

    window.saveEditConsumer = function() {
        var id = document.getElementById('editConsId').value;
        var c = cachedConsumers.find(function(x) { return x.id === id; });
        var queueTopic = document.getElementById('editConsQueue').value;
        var req = {
            name: document.getElementById('editConsName').value,
            backend: c ? c.backend : 'rabbitmq',
            consumerGroup: document.getElementById('editConsGroup').value,
            config: {
                prefetchCount: parseInt(document.getElementById('editConsPrefetch').value) || 10,
                maxRetries: parseInt(document.getElementById('editConsMaxRetries').value) || 3,
                dlqEnabled: document.getElementById('editConsDLQ').checked
            }
        };
        if (c && c.backend === 'kafka') {
            req.topic = queueTopic;
        } else {
            req.queue = queueTopic;
        }
        req.exchange = document.getElementById('editConsExchange').value;
        req.routingKey = document.getElementById('editConsRoutingKey').value;
        apiPatch('/consumers/' + id, req).then(function() {
            closeModal('editConsumerModal');
            loadConsumers();
            loadStats();
        }).catch(function(e) { alert('Error: ' + e); });
    };

    window.showCreateConsumerModal = function() {
        document.getElementById('createConsumerModal').style.display = 'flex';
    };

    window.createConsumer = function() {
        var backend = document.getElementById('consBackend').value;
        var req = {
            name: document.getElementById('consName').value,
            backend: backend,
            consumerGroup: document.getElementById('consGroup').value,
            config: {
                prefetchCount: parseInt(document.getElementById('consPrefetch').value) || 10,
                maxRetries: parseInt(document.getElementById('consMaxRetries').value) || 3,
                dlqEnabled: document.getElementById('consDLQ').checked
            }
        };
        var queueTopic = document.getElementById('consQueue').value;
        if (backend === 'kafka') {
            req.topic = queueTopic;
        } else {
            req.queue = queueTopic;
        }
        req.exchange = document.getElementById('consExchange').value;
        req.routingKey = document.getElementById('consRoutingKey').value;
        apiPost('/consumers', req).then(function() {
            closeModal('createConsumerModal');
            loadConsumers();
            loadStats();
        }).catch(function(e) { alert('Error: ' + e); });
    };

    // ---- Messages ----
    window.loadMessages = function() {
        apiGet('/messages?limit=100').then(function(data) {
            var msgs = data.messages || [];
            var tbody = document.getElementById('messagesBody');
            if (!msgs.length) {
                tbody.innerHTML = '<tr><td colspan="6" class="empty-row">No messages yet</td></tr>';
                return;
            }
            tbody.innerHTML = msgs.map(function(m) {
                return '<tr>' +
                    '<td><small>' + (m.id || '-') + '</small></td>' +
                    '<td>' + (m.producerId || '-') + '</td>' +
                    '<td>' + statusBadge(m.status) + '</td>' +
                    '<td>' + (m.contentType || '-') + '</td>' +
                    '<td>' + formatTime(m.timestamp) + '</td>' +
                    '<td>' + jsonPreview(m.body) + '</td>' +
                    '</tr>';
            }).join('');
        }).catch(function() {});
    };

    // ---- DLQ ----
    window.loadDLQ = function() {
        apiGet('/dlq').then(function(data) {
            var entries = data.dlq || [];
            var tbody = document.getElementById('dlqBody');
            if (!entries.length) {
                tbody.innerHTML = '<tr><td colspan="7" class="empty-row">No dead-lettered messages</td></tr>';
                return;
            }
            tbody.innerHTML = entries.map(function(e) {
                var replayBtn = e.replayed ? '<span style="color:var(--text-muted)">Replayed</span>' :
                    '<button class="action-btn" onclick="window._conductorReplayDLQ(\'' + e.id + '\')">Replay</button>';
                return '<tr>' +
                    '<td><small>' + e.id + '</small></td>' +
                    '<td>' + (e.consumerId || '-') + '</td>' +
                    '<td>' + (e.originalQueue || '-') + '</td>' +
                    '<td style="color:#ff4d4d">' + (e.errorMessage || '-') + '</td>' +
                    '<td>' + (e.retryCount || 0) + '</td>' +
                    '<td>' + formatTime(e.deadLetteredAt) + '</td>' +
                    '<td>' + replayBtn + '</td>' +
                    '</tr>';
            }).join('');
        }).catch(function() {});
    };

    window._conductorReplayDLQ = function(id) {
        apiPost('/dlq/' + id + '/replay', {}).then(function() { loadDLQ(); loadStats(); });
    };

    // ---- Publish ----
    window.publishMessage = function() {
        var prodId = document.getElementById('publishProducer').value;
        if (!prodId) { alert('Select a producer first'); return; }
        var bodyText = document.getElementById('publishBody').value;
        var body;
        try { body = JSON.parse(bodyText); } catch(e) { alert('Invalid JSON body'); return; }
        var req = {
            producerId: prodId,
            body: body,
            routingKey: document.getElementById('publishRoutingKey').value,
            correlationId: document.getElementById('publishCorrelationId').value
        };
        var resultEl = document.getElementById('publishResult');
        apiPost('/publish', req).then(function(data) {
            if (data.error) {
                resultEl.className = 'publish-result error';
                resultEl.textContent = 'Error: ' + data.error;
            } else {
                resultEl.className = 'publish-result success';
                resultEl.textContent = 'Published! Message ID: ' + (data.id || 'unknown');
                loadStats();
            }
        }).catch(function(e) {
            resultEl.className = 'publish-result error';
            resultEl.textContent = 'Error: ' + e;
        });
    };

    // ---- Live Stream ----
    window.toggleStream = function() {
        if (streamConnected) {
            disconnectStream();
        } else {
            connectStream();
        }
    };

    function connectStream() {
        var base = window.resolveBackendURL();
        var token = localStorage.getItem('auth_token') || '';
        wsURL = base.replace(/^http/, 'ws') + '/ws/conductor' + (token ? '?token=' + encodeURIComponent(token) : '');
        try {
            streamWS = new WebSocket(wsURL);
        } catch(e) {
            // fallback to SSE
            connectSSE();
            return;
        }

        streamWS.onopen = function() {
            streamConnected = true;
            var indicator = document.getElementById('streamIndicator');
            indicator.textContent = '● Connected';
            indicator.className = 'stream-indicator connected';
            document.getElementById('streamToggle').textContent = 'Disconnect';
            document.getElementById('streamContainer').innerHTML = '';
        };

        streamWS.onmessage = function(event) {
            try {
                var data = JSON.parse(event.data);
                renderStreamMessages(data.messages || []);
                if (data.stats) {
                    document.getElementById('statSent').textContent = data.stats.totalSent || 0;
                    document.getElementById('statReceived').textContent = data.stats.totalReceived || 0;
                    document.getElementById('statAcked').textContent = data.stats.totalAcked || 0;
                    document.getElementById('statFailed').textContent = data.stats.totalFailed || 0;
                }
            } catch(e) {}
        };

        streamWS.onclose = function() { disconnectStream(); };
        streamWS.onerror = function() { disconnectStream(); connectSSE(); };
    }

    function connectSSE() {
        streamConnected = true;
        var indicator = document.getElementById('streamIndicator');
        indicator.textContent = '● Connected (SSE)';
        indicator.className = 'stream-indicator connected';
        document.getElementById('streamToggle').textContent = 'Disconnect';
        document.getElementById('streamContainer').innerHTML = '';

        var es = new EventSource(API + '/stream' + (localStorage.getItem('auth_token') ? '?token=' + encodeURIComponent(localStorage.getItem('auth_token')) : ''));
        window._conductorSSE = es;
        es.onmessage = function(event) {
            try {
                var msgs = JSON.parse(event.data);
                renderStreamMessages(msgs);
            } catch(e) {}
        };
        es.onerror = function() { disconnectStream(); };
    }

    function disconnectStream() {
        streamConnected = false;
        if (streamWS) { try { streamWS.close(); } catch(e){} streamWS = null; }
        if (window._conductorSSE) { try { window._conductorSSE.close(); } catch(e){} window._conductorSSE = null; }
        var indicator = document.getElementById('streamIndicator');
        indicator.textContent = '● Disconnected';
        indicator.className = 'stream-indicator disconnected';
        document.getElementById('streamToggle').textContent = 'Connect';
    }

    function renderStreamMessages(msgs) {
        var container = document.getElementById('streamContainer');
        container.innerHTML = msgs.map(function(m) {
            var time = formatTime(m.timestamp);
            return '<div class="stream-msg">' +
                '<span class="stream-msg-time">' + time + '</span>' +
                '<span class="stream-msg-status">' + statusBadge(m.status) + '</span>' +
                '<span class="stream-msg-body">' + JSON.stringify(m.body || {}) + '</span>' +
                '</div>';
        }).join('');
        container.scrollTop = container.scrollHeight;
    }

    // ---- Hover Metrics Tooltip ----
    function bindHoverMetrics() {
        var tooltip = document.getElementById('metricsTooltip');
        if (!tooltip) return;
        document.querySelectorAll('.hoverable-row').forEach(function(row) {
            row.addEventListener('mouseenter', function(e) {
                var type = row.getAttribute('data-type');
                var id = row.getAttribute('data-id');
                var html = '';
                if (type === 'producer') {
                    var p = cachedProducers.find(function(x) { return x.id === id; });
                    if (!p) return;
                    html = '<div class="tt-title">' + p.name + '</div>' +
                        '<div class="tt-row"><span class="tt-label">Status</span><span class="tt-value">' + (p.status || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Backend</span><span class="tt-value">' + (p.backend || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Messages Sent</span><span class="tt-value">' + (p.messagesSent || 0) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Last Sent</span><span class="tt-value">' + formatTime(p.lastSentAt) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Content Type</span><span class="tt-value">' + (p.contentType || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Created</span><span class="tt-value">' + formatTime(p.createdAt) + '</span></div>';
                } else if (type === 'consumer') {
                    var c = cachedConsumers.find(function(x) { return x.id === id; });
                    if (!c) return;
                    html = '<div class="tt-title">' + c.name + '</div>' +
                        '<div class="tt-row"><span class="tt-label">Status</span><span class="tt-value">' + (c.status || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Backend</span><span class="tt-value">' + (c.backend || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Received</span><span class="tt-value">' + (c.messagesReceived || 0) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Acked</span><span class="tt-value">' + (c.messagesAcked || 0) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Failed</span><span class="tt-value">' + (c.messagesFailed || 0) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Last Received</span><span class="tt-value">' + formatTime(c.lastReceivedAt) + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">Prefetch</span><span class="tt-value">' + ((c.config && c.config.prefetchCount) || '-') + '</span></div>' +
                        '<div class="tt-row"><span class="tt-label">DLQ</span><span class="tt-value">' + ((c.config && c.config.dlqEnabled) ? 'Enabled' : 'Disabled') + '</span></div>';
                }
                tooltip.innerHTML = html;
                tooltip.style.display = 'block';
            });
            row.addEventListener('mousemove', function(e) {
                tooltip.style.left = (e.clientX + 16) + 'px';
                tooltip.style.top = (e.clientY + 10) + 'px';
            });
            row.addEventListener('mouseleave', function() {
                tooltip.style.display = 'none';
            });
        });
    }

    // ---- Tabs ----
    window.switchTab = function(tab) {
        document.querySelectorAll('.tab-content').forEach(function(el) { el.classList.remove('active'); });
        document.querySelectorAll('.tab-btn').forEach(function(el) { el.classList.remove('active'); });
        document.getElementById('tab-' + tab).classList.add('active');
        document.querySelector('.tab-btn[data-tab="' + tab + '"]').classList.add('active');

        if (tab === 'producers') loadProducers();
        if (tab === 'consumers') loadConsumers();
        if (tab === 'messages') window.loadMessages();
        if (tab === 'dlq') window.loadDLQ();
    };

    window.closeModal = function(id) {
        document.getElementById(id).style.display = 'none';
    };

    // Initial load
    loadStats();
    loadProducers();
    setInterval(loadStats, 5000);
})();
