(function() {
    'use strict';

    var API = window.resolveBackendURL() + '/api/v1/conductor';
    var wsURL = '';
    var streamWS = null;
    var streamConnected = false;

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
            var tbody = document.getElementById('producersBody');
            if (!prods.length) {
                tbody.innerHTML = '<tr><td colspan="8" class="empty-row">No producers configured yet</td></tr>';
                return;
            }
            tbody.innerHTML = prods.map(function(p) {
                var target = p.backend === 'kafka' ? (p.topic || '-') : (p.exchange || '-');
                return '<tr>' +
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
        }).catch(function() {});
    }

    function producerActions(p) {
        var html = '';
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
            var tbody = document.getElementById('consumersBody');
            if (!cons.length) {
                tbody.innerHTML = '<tr><td colspan="9" class="empty-row">No consumers configured yet</td></tr>';
                return;
            }
            tbody.innerHTML = cons.map(function(c) {
                var target = c.backend === 'kafka' ? (c.topic || '-') : (c.queue || '-');
                return '<tr>' +
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
        }).catch(function() {});
    }

    function consumerActions(c) {
        var html = '';
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
