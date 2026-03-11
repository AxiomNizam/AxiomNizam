// =====================================================
// CDC / ETL Dashboard — Frontend Logic
// =====================================================

(function() {
    'use strict';

    const API = window.BACKEND_URL || 'http://localhost:8000';
    let currentTab = 'etl';
    let createType = 'etl';
    let chartInstances = {};

    // --- Tab Switching ---
    window.switchTab = function(tab) {
        currentTab = tab;
        document.querySelectorAll('.cdc-etl-tab').forEach(t => t.classList.remove('active'));
        document.querySelectorAll('.panel-tab-content').forEach(p => p.classList.remove('active'));
        document.getElementById('tab-' + (tab === 'observability' ? 'obs' : tab === 'connectors' ? 'conn' : tab)).classList.add('active');
        document.getElementById('panel-' + tab).classList.add('active');
    };

    // --- Data Fetching ---
    async function fetchJSON(path) {
        try {
            const resp = await fetch(API + path);
            if (!resp.ok) throw new Error('HTTP ' + resp.status);
            return await resp.json();
        } catch (e) {
            console.error('Fetch error:', path, e);
            return null;
        }
    }

    // --- Number formatting ---
    function fmtNum(n) {
        if (n == null) return '0';
        if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
        if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
        return String(n);
    }

    function fmtTime(ts) {
        if (!ts) return '—';
        const d = new Date(ts);
        return d.toLocaleString(undefined, { month:'short', day:'numeric', hour:'2-digit', minute:'2-digit' });
    }

    // --- Status Badge ---
    function statusBadge(s) {
        return '<span class="status-badge ' + s + '"><span class="status-dot"></span> ' + s + '</span>';
    }

    function tagList(tags) {
        if (!tags || !tags.length) return '—';
        return '<div class="tag-list">' + tags.map(t => '<span class="tag">' + escapeHtml(t) + '</span>').join('') + '</div>';
    }

    function escapeHtml(str) {
        const div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    // ==============================
    // ETL
    // ==============================
    async function loadETL() {
        const [pData, rData, oData] = await Promise.all([
            fetchJSON('/api/v1/etl/pipelines'),
            fetchJSON('/api/v1/etl/runs'),
            fetchJSON('/api/v1/etl/observability')
        ]);

        if (pData) renderETLPipelines(pData.pipelines || []);
        if (rData) renderETLRuns(rData.runs || []);
        if (oData) renderETLKPIs(oData.observability);
        if (pData) document.getElementById('etl-count').textContent = (pData.pipelines || []).length;
    }

    function renderETLKPIs(obs) {
        if (!obs) return;
        const el = document.getElementById('etl-kpis');
        el.innerHTML = [
            kpiCard('Pipelines', obs.pipelines_total, '', ''),
            kpiCard('Total Runs', obs.runs_total, '', 'info'),
            kpiCard('Successful', obs.runs_success, '', 'success'),
            kpiCard('Failed', obs.runs_failed, '', 'danger'),
            kpiCard('Running', obs.runs_running, '', 'warning'),
            kpiCard('Rows Read', fmtNum(obs.total_rows_read), '', 'info'),
            kpiCard('Rows Written', fmtNum(obs.total_rows_written), '', 'success'),
            kpiCard('Avg Duration', obs.avg_duration_seconds ? obs.avg_duration_seconds.toFixed(1) + 's' : '—', '', ''),
        ].join('');
    }

    function kpiCard(label, value, sub, cls) {
        return '<div class="kpi-card ' + (cls ? 'kpi-' + cls : '') + '">' +
            '<div class="kpi-label">' + label + '</div>' +
            '<div class="kpi-value">' + value + '</div>' +
            (sub ? '<div class="kpi-sub">' + sub + '</div>' : '') +
            '</div>';
    }

    function renderETLPipelines(pipelines) {
        const tbody = document.getElementById('etl-pipeline-body');
        if (!pipelines.length) {
            tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:30px;color:#888">No ETL pipelines found</td></tr>';
            return;
        }
        tbody.innerHTML = pipelines.map(p => '<tr>' +
            '<td><strong>' + escapeHtml(p.name) + '</strong><br><small style="color:#888">' + escapeHtml(p.id) + '</small></td>' +
            '<td>' + statusBadge(p.status) + '</td>' +
            '<td>' + (p.steps ? p.steps.length : 0) + ' steps</td>' +
            '<td>' + (p.schedule || '—') + '</td>' +
            '<td>' + (p.run_count || 0) + '</td>' +
            '<td>' + fmtTime(p.last_run_at) + '</td>' +
            '<td>' + tagList(p.tags) + '</td>' +
            '<td class="action-btns">' +
                '<button class="action-btn" onclick="runETLPipeline(\'' + p.id + '\')" title="Run">▶</button>' +
                '<button class="action-btn danger" onclick="deleteETLPipeline(\'' + p.id + '\')" title="Delete">🗑</button>' +
            '</td>' +
        '</tr>').join('');
    }

    function renderETLRuns(runs) {
        const tbody = document.getElementById('etl-runs-body');
        if (!runs.length) {
            tbody.innerHTML = '<tr><td colspan="8" style="text-align:center;padding:30px;color:#888">No runs yet</td></tr>';
            return;
        }
        // Sort newest first
        runs.sort((a, b) => new Date(b.started_at) - new Date(a.started_at));
        tbody.innerHTML = runs.slice(0, 20).map(r => '<tr>' +
            '<td><small>' + escapeHtml(r.id) + '</small></td>' +
            '<td>' + escapeHtml(r.pipeline_id) + '</td>' +
            '<td>' + statusBadge(r.status) + '</td>' +
            '<td>' + (r.trigger || '—') + '</td>' +
            '<td>' + fmtNum(r.rows_read) + '</td>' +
            '<td>' + fmtNum(r.rows_written) + '</td>' +
            '<td>' + (r.duration || '—') + '</td>' +
            '<td>' + fmtTime(r.started_at) + '</td>' +
        '</tr>').join('');
    }

    window.runETLPipeline = async function(id) {
        await fetch(API + '/api/v1/etl/pipelines/' + encodeURIComponent(id) + '/run', { method: 'POST' });
        setTimeout(loadETL, 500);
    };

    window.deleteETLPipeline = async function(id) {
        if (!confirm('Delete pipeline ' + id + '?')) return;
        await fetch(API + '/api/v1/etl/pipelines/' + encodeURIComponent(id), { method: 'DELETE' });
        loadETL();
    };

    // ==============================
    // CDC
    // ==============================
    async function loadCDC() {
        const [pData, oData] = await Promise.all([
            fetchJSON('/api/v1/cdc/pipelines'),
            fetchJSON('/api/v1/cdc/observability')
        ]);

        if (pData) renderCDCPipelines(pData.pipelines || []);
        if (oData) renderCDCKPIs(oData.observability);
        if (pData) document.getElementById('cdc-count').textContent = (pData.pipelines || []).length;
    }

    function renderCDCKPIs(obs) {
        if (!obs) return;
        const el = document.getElementById('cdc-kpis');
        el.innerHTML = [
            kpiCard('Total Pipelines', obs.pipelines_total, '', ''),
            kpiCard('Active', obs.pipelines_active, '', 'success'),
            kpiCard('Paused', obs.pipelines_paused, '', 'warning'),
            kpiCard('Failed', obs.pipelines_failed, '', 'danger'),
            kpiCard('Total Events', fmtNum(obs.total_events), '', 'info'),
            kpiCard('Events/sec', obs.events_per_second ? obs.events_per_second.toFixed(1) : '0', '', 'success'),
            kpiCard('Errors', fmtNum(obs.total_errors), '', 'danger'),
            kpiCard('Avg Lag', obs.avg_lag_ms ? obs.avg_lag_ms.toFixed(0) + 'ms' : '—', '', ''),
        ].join('');
    }

    function renderCDCPipelines(pipelines) {
        const tbody = document.getElementById('cdc-pipeline-body');
        if (!pipelines.length) {
            tbody.innerHTML = '<tr><td colspan="9" style="text-align:center;padding:30px;color:#888">No CDC pipelines found</td></tr>';
            return;
        }
        tbody.innerHTML = pipelines.map(p => '<tr>' +
            '<td><strong>' + escapeHtml(p.name) + '</strong><br><small style="color:#888">' + escapeHtml(p.id) + '</small></td>' +
            '<td>' + statusBadge(p.status) + '</td>' +
            '<td>' + (p.source ? escapeHtml(p.source.type) : '—') + '</td>' +
            '<td>' + (p.sink ? escapeHtml(p.sink.type) : '—') + '</td>' +
            '<td>' + fmtNum(p.event_count) + '</td>' +
            '<td>' + fmtNum(p.error_count) + '</td>' +
            '<td>' + (p.lag || '—') + '</td>' +
            '<td>' + tagList(p.tags) + '</td>' +
            '<td class="action-btns">' +
                (p.status !== 'active' ? '<button class="action-btn" onclick="cdcAction(\'' + p.id + '\',\'start\')" title="Start">▶</button>' : '') +
                (p.status === 'active' ? '<button class="action-btn" onclick="cdcAction(\'' + p.id + '\',\'pause\')" title="Pause">⏸</button>' : '') +
                (p.status !== 'stopped' ? '<button class="action-btn" onclick="cdcAction(\'' + p.id + '\',\'stop\')" title="Stop">⏹</button>' : '') +
                '<button class="action-btn danger" onclick="deleteCDCPipeline(\'' + p.id + '\')" title="Delete">🗑</button>' +
            '</td>' +
        '</tr>').join('');
    }

    window.cdcAction = async function(id, action) {
        await fetch(API + '/api/v1/cdc/pipelines/' + encodeURIComponent(id) + '/' + action, { method: 'POST' });
        loadCDC();
    };

    window.deleteCDCPipeline = async function(id) {
        if (!confirm('Delete CDC pipeline ' + id + '?')) return;
        await fetch(API + '/api/v1/cdc/pipelines/' + encodeURIComponent(id), { method: 'DELETE' });
        loadCDC();
    };

    // ==============================
    // Observability Charts
    // ==============================
    async function loadObservability() {
        const [etlObs, cdcObs] = await Promise.all([
            fetchJSON('/api/v1/etl/observability'),
            fetchJSON('/api/v1/cdc/observability')
        ]);

        const eObs = etlObs ? etlObs.observability : {};
        const cObs = cdcObs ? cdcObs.observability : {};

        // Combined KPIs
        const el = document.getElementById('obs-kpis');
        el.innerHTML = [
            kpiCard('ETL Pipelines', eObs.pipelines_total || 0, '', ''),
            kpiCard('CDC Pipelines', cObs.pipelines_total || 0, '', ''),
            kpiCard('ETL Runs', eObs.runs_total || 0, 'Success: ' + (eObs.runs_success || 0), 'info'),
            kpiCard('CDC Events', fmtNum(cObs.total_events), fmtNum(cObs.events_per_second) + ' evt/s', 'success'),
            kpiCard('ETL Rows', fmtNum((eObs.total_rows_read || 0) + (eObs.total_rows_written || 0)), '', 'info'),
            kpiCard('CDC Errors', fmtNum(cObs.total_errors), (cObs.error_rate_percent || 0).toFixed(2) + '% rate', 'danger'),
        ].join('');

        // Charts
        renderETLRunsChart(eObs);
        renderCDCThroughputChart(cObs);
        renderETLStepsChart(eObs);
        renderCDCOpsChart(cObs);

        // Error lists
        renderErrorList('etl-error-list', eObs.errors_by_type || {});
        renderErrorList('cdc-error-list', cObs.errors_by_type || {});
    }

    function renderErrorList(elId, errors) {
        const el = document.getElementById(elId);
        const entries = Object.entries(errors);
        if (!entries.length) {
            el.innerHTML = '<li style="color:#888">No errors recorded</li>';
            return;
        }
        el.innerHTML = entries.map(([k, v]) => '<li><span class="error-type">' + escapeHtml(k) + '</span><span class="error-count">' + v + '</span></li>').join('');
    }

    function destroyChart(id) {
        if (chartInstances[id]) {
            chartInstances[id].destroy();
            delete chartInstances[id];
        }
    }

    function renderETLRunsChart(obs) {
        destroyChart('chart-etl-runs');
        const ctx = document.getElementById('chart-etl-runs');
        if (!ctx) return;

        const log = obs.throughput_log || [];
        const labels = log.map(p => {
            const d = new Date(p.timestamp);
            return d.getHours() + ':' + String(d.getMinutes()).padStart(2, '0');
        });

        chartInstances['chart-etl-runs'] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels.length ? labels : ['No data'],
                datasets: [{
                    label: 'Rows/sec',
                    data: log.map(p => p.rows_per_sec || 0),
                    borderColor: '#4f46e5',
                    backgroundColor: 'rgba(79,70,229,0.1)',
                    fill: true,
                    tension: 0.3
                }]
            },
            options: { responsive: true, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true } } }
        });
    }

    function renderCDCThroughputChart(obs) {
        destroyChart('chart-cdc-throughput');
        const ctx = document.getElementById('chart-cdc-throughput');
        if (!ctx) return;

        const log = obs.throughput_log || [];
        const labels = log.map(p => {
            const d = new Date(p.timestamp);
            return d.getHours() + ':' + String(d.getMinutes()).padStart(2, '0');
        });

        chartInstances['chart-cdc-throughput'] = new Chart(ctx, {
            type: 'line',
            data: {
                labels: labels.length ? labels : ['No data'],
                datasets: [{
                    label: 'Events/sec',
                    data: log.map(p => p.events_per_sec || 0),
                    borderColor: '#10b981',
                    backgroundColor: 'rgba(16,185,129,0.1)',
                    fill: true,
                    tension: 0.3
                }, {
                    label: 'Lag (ms)',
                    data: log.map(p => p.lag_ms || 0),
                    borderColor: '#f59e0b',
                    backgroundColor: 'rgba(245,158,11,0.1)',
                    fill: false,
                    tension: 0.3,
                    yAxisID: 'y1'
                }]
            },
            options: {
                responsive: true,
                scales: {
                    y: { beginAtZero: true, position: 'left', title: { display: true, text: 'Events/sec' } },
                    y1: { beginAtZero: true, position: 'right', title: { display: true, text: 'Lag (ms)' }, grid: { drawOnChartArea: false } }
                }
            }
        });
    }

    function renderETLStepsChart(obs) {
        destroyChart('chart-etl-steps');
        const ctx = document.getElementById('chart-etl-steps');
        if (!ctx) return;

        const stats = obs.step_type_stats || {};
        const labels = Object.keys(stats);
        const data = Object.values(stats);
        const colors = ['#4f46e5', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#14b8a6', '#f97316', '#6366f1'];

        chartInstances['chart-etl-steps'] = new Chart(ctx, {
            type: 'doughnut',
            data: {
                labels: labels.length ? labels : ['No data'],
                datasets: [{ data: data.length ? data : [1], backgroundColor: colors }]
            },
            options: { responsive: true, plugins: { legend: { position: 'right' } } }
        });
    }

    function renderCDCOpsChart(obs) {
        destroyChart('chart-cdc-ops');
        const ctx = document.getElementById('chart-cdc-ops');
        if (!ctx) return;

        const events = obs.events_by_operation || {};
        chartInstances['chart-cdc-ops'] = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: Object.keys(events).length ? Object.keys(events) : ['No data'],
                datasets: [{
                    label: 'Events',
                    data: Object.values(events).length ? Object.values(events) : [0],
                    backgroundColor: ['#10b981', '#3b82f6', '#ef4444', '#f59e0b']
                }]
            },
            options: { responsive: true, plugins: { legend: { display: false } }, scales: { y: { beginAtZero: true } } }
        });
    }

    // ==============================
    // Connectors
    // ==============================
    async function loadConnectors() {
        const [etlConn, cdcSrc, cdcSnk] = await Promise.all([
            fetchJSON('/api/v1/etl/connectors'),
            fetchJSON('/api/v1/cdc/sources'),
            fetchJSON('/api/v1/cdc/sinks')
        ]);

        if (etlConn) renderConnectors('etl-connectors', etlConn.connectors || []);
        if (cdcSrc) renderConnectors('cdc-sources', cdcSrc.sources || []);
        if (cdcSnk) renderConnectors('cdc-sinks', cdcSnk.sinks || []);
    }

    function renderConnectors(elId, list) {
        const el = document.getElementById(elId);
        if (!list.length) {
            el.innerHTML = '<p style="color:#888">No connectors available</p>';
            return;
        }
        el.innerHTML = list.map(c => '<div class="connector-card">' +
            '<div class="connector-icon">' + (c.icon || '🔌') + '</div>' +
            '<div class="connector-name">' + escapeHtml(c.name) + '</div>' +
            '<div class="connector-type">' + escapeHtml(c.category || c.method || c.id) + '</div>' +
        '</div>').join('');
    }

    // ==============================
    // Create Modal
    // ==============================
    window.openCreateModal = function(type) {
        createType = type;
        document.getElementById('modal-title').textContent = type === 'etl' ? 'Create ETL Pipeline' : 'Create CDC Pipeline';
        document.getElementById('modal-schedule-group').style.display = type === 'etl' ? 'block' : 'none';
        document.getElementById('modal-source-group').style.display = type === 'cdc' ? 'block' : 'none';
        document.getElementById('modal-sink-group').style.display = type === 'cdc' ? 'block' : 'none';
        document.getElementById('modal-name').value = '';
        document.getElementById('modal-desc').value = '';
        document.getElementById('modal-schedule').value = '';
        document.getElementById('createModal').classList.add('visible');
    };

    window.closeCreateModal = function() {
        document.getElementById('createModal').classList.remove('visible');
    };

    window.submitCreate = async function() {
        const name = document.getElementById('modal-name').value.trim();
        const desc = document.getElementById('modal-desc').value.trim();
        if (!name) { alert('Name is required'); return; }

        let body;
        let url;

        if (createType === 'etl') {
            body = { name: name, description: desc, schedule: document.getElementById('modal-schedule').value.trim(), steps: [], tags: [] };
            url = '/api/v1/etl/pipelines';
        } else {
            const srcType = document.getElementById('modal-source-type').value;
            const sinkType = document.getElementById('modal-sink-type').value;
            body = {
                name: name,
                description: desc,
                source: { type: srcType, connector: srcType, config: {} },
                sink: { type: sinkType, connector: sinkType, config: {} },
                filters: {},
                tags: []
            };
            url = '/api/v1/cdc/pipelines';
        }

        await fetch(API + url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(body)
        });

        closeCreateModal();
        if (createType === 'etl') loadETL(); else loadCDC();
    };

    // Populate CDC source/sink dropdowns
    async function initDropdowns() {
        const [srcData, sinkData] = await Promise.all([
            fetchJSON('/api/v1/cdc/sources'),
            fetchJSON('/api/v1/cdc/sinks')
        ]);

        const srcSel = document.getElementById('modal-source-type');
        const sinkSel = document.getElementById('modal-sink-type');

        if (srcData && srcData.sources) {
            srcSel.innerHTML = srcData.sources.map(s => '<option value="' + s.id + '">' + s.icon + ' ' + escapeHtml(s.name) + '</option>').join('');
        }
        if (sinkData && sinkData.sinks) {
            sinkSel.innerHTML = sinkData.sinks.map(s => '<option value="' + s.id + '">' + s.icon + ' ' + escapeHtml(s.name) + '</option>').join('');
        }
    }

    // ==============================
    // Init
    // ==============================
    async function init() {
        await Promise.all([loadETL(), loadCDC(), loadObservability(), loadConnectors(), initDropdowns()]);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
