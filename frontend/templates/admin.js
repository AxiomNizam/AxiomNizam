// Admin Dashboard JS — API Builder, File-to-Dashboard, Dashboard↔GIS Converter, File Scanner
const BACKEND_URL = (() => {
    const elem = document.getElementById('backendURL');
    if (elem && elem.textContent) {
        let url = elem.textContent.trim();
        if (url.includes('axiomnizam:8000')) {
            url = url.replace('axiomnizam:8000', 'localhost:8000');
        }
        return url;
    }
    return 'http://' + window.location.hostname + ':8000';
})();

let filteredMethod = 'ALL';
let currentCSVUploadId = null;
let currentDashMappings = [];
let builderDataServers = [];

const DEFAULT_BUILDER_DATABASES = ['mysql', 'postgres', 'mariadb', 'percona', 'oracle'];

console.log('Admin - Backend URL:', BACKEND_URL);

function getAuthHeaders() {
    const token = localStorage.getItem('authToken');
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;
    return headers;
}

function fetchJSON(path) {
    return fetch(BACKEND_URL + path, { headers: getAuthHeaders() }).then(function(r) { return r.json(); });
}
function postJSON(path, body) {
    return fetch(BACKEND_URL + path, { method: 'POST', headers: getAuthHeaders(), body: JSON.stringify(body) }).then(function(r) { return r.json(); });
}
function putJSON(path, body) {
    return fetch(BACKEND_URL + path, { method: 'PUT', headers: getAuthHeaders(), body: JSON.stringify(body) }).then(function(r) { return r.json(); });
}
function deleteJSON(path) {
    return fetch(BACKEND_URL + path, { method: 'DELETE', headers: getAuthHeaders() }).then(function(r) { return r.json(); });
}

// ===================================================================
// Tab switching & init
// ===================================================================
window.addEventListener('DOMContentLoaded', function() {
    var userName = localStorage.getItem('userName');
    if (userName) {
        var el = document.getElementById('adminUserName');
        if (el) el.textContent = userName;
    }
    loadBuilderSummary();
    loadCustomAPIs();
    loadBuilderDataSources();
    loadCSVHistory();
    loadAPIs();
    setupCSVDropZone();
    setupScanDropZone();
    loadScannerHealth();
    applyRoleRestrictions();
});

function switchTab(tabName) {
    var tabs = document.querySelectorAll('.tab-content');
    tabs.forEach(function(t) { t.classList.remove('active'); });
    var btns = document.querySelectorAll('.tab-btn');
    btns.forEach(function(b) { b.classList.remove('active'); });
    var sel = document.getElementById(tabName);
    if (sel) sel.classList.add('active');
    if (event && event.currentTarget) event.currentTarget.classList.add('active');

    if (tabName === 'api-builder') { loadBuilderSummary(); loadCustomAPIs(); }
    if (tabName === 'csv-upload') { loadCSVHistory(); }
    if (tabName === 'converter') { loadConverterDropdowns(); loadConversionHistory(); }
    if (tabName === 'file-scanner') { loadScanHistory(); loadScannerHealth(); }
}

// ===================================================================
// API Builder
// ===================================================================
function loadBuilderSummary() {
    fetchJSON('/api/v1/builder/summary').then(function(d) {
        var el = document.getElementById('builderSummary');
        if (!el) return;
        el.innerHTML =
            '<div class="summary-card"><div class="sc-value">' + (d.total_apis || 0) + '</div><div class="sc-label">Total APIs</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.active || 0) + '</div><div class="sc-label">Active</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.draft || 0) + '</div><div class="sc-label">Draft</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.total_hits || 0) + '</div><div class="sc-label">Total Hits</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.total_csv_uploads || 0) + '</div><div class="sc-label">CSV Uploads</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.total_conversions || 0) + '</div><div class="sc-label">Conversions</div></div>';
    }).catch(function() {});
}

function loadCustomAPIs() {
    var catEl = document.getElementById('apiCategoryFilter');
    var statEl = document.getElementById('apiStatusFilter');
    var cat = catEl ? catEl.value : '';
    var status = statEl ? statEl.value : '';
    var q = '/api/v1/builder/apis?';
    if (cat) q += 'category=' + encodeURIComponent(cat) + '&';
    if (status) q += 'status=' + encodeURIComponent(status);

    fetchJSON(q).then(function(d) {
        var list = d.apis || [];
        var el = document.getElementById('apiBuilderList');
        if (!el) return;
        if (list.length === 0) {
            el.innerHTML = '<div class="empty-state">No custom APIs found. Click <strong>+ Create API</strong> to get started.</div>';
            return;
        }
        var html = '<table class="admin-table"><thead><tr>' +
            '<th>Method</th><th>Name</th><th>Path</th><th>Category</th><th>Source DB</th><th>Source Server</th><th>Status</th><th>Hits</th><th>Actions</th>' +
            '</tr></thead><tbody>';
        list.forEach(function(api) {
            var safeId = api.id.replace(/'/g, "\\'");
            var actionsHtml = '<button class="btn-sm btn-test" onclick="testCustomAPI(\'' + safeId + '\')">Test</button> ';
            if (canModify()) {
                actionsHtml += '<button class="btn-sm btn-del" onclick="deleteCustomAPI(\'' + safeId + '\')">Del</button>';
            }
            html += '<tr>' +
                '<td><span class="method-badge method-' + api.method.toLowerCase() + '">' + api.method + '</span></td>' +
                '<td>' + escapeHtml(api.name) + '</td>' +
                '<td><code>' + escapeHtml(api.path) + '</code></td>' +
                '<td>' + escapeHtml(api.category || '-') + '</td>' +
                '<td>' + escapeHtml(api.source_database || '-') + '</td>' +
                '<td>' + escapeHtml(api.source_server || 'default') + '</td>' +
                '<td><span class="status-badge status-' + api.status + '">' + api.status + '</span></td>' +
                '<td>' + (api.hit_count || 0) + '</td>' +
                '<td>' + actionsHtml + '</td></tr>';
        });
        html += '</tbody></table>';
        el.innerHTML = html;
    }).catch(function() {
        var el = document.getElementById('apiBuilderList');
        if (el) el.innerHTML = '<div class="error-state">Failed to load APIs</div>';
    });
}

function escapeHtml(str) {
    var div = document.createElement('div');
    div.textContent = str;
    return div.innerHTML;
}

function closeCreateAPIModal() {
    document.getElementById('createAPIModal').style.display = 'none';
    document.getElementById('createAPIForm').reset();
    var ttlGroup = document.getElementById('cacheTTLGroup');
    if (ttlGroup) ttlGroup.style.display = 'none';
    updateBuilderSourceServers();
}

function toggleCacheTTL() {
    var checked = document.getElementById('apiCacheInput').checked;
    var group = document.getElementById('cacheTTLGroup');
    if (group) group.style.display = checked ? 'block' : 'none';
}

function submitCreateAPI(e) {
    e.preventDefault();
    var mockRaw = document.getElementById('apiMockResponseInput').value.trim();
    var mockResp = null;
    if (mockRaw) {
        try { mockResp = JSON.parse(mockRaw); } catch(err) { alert('Invalid JSON in mock response'); return; }
    }
    var qpRaw = document.getElementById('apiQueryParamsInput').value.trim();
    var queryParams = [];
    if (qpRaw) {
        qpRaw.split('\n').forEach(function(line) {
            var parts = line.trim().split(':');
            if (parts.length >= 2) {
                queryParams.push({ name: parts[0].trim(), type: parts[1].trim(), required: parts.length > 2 && parts[2].trim() === 'true' });
            }
        });
    }

    var body = {
        name: document.getElementById('apiNameInput').value,
        method: document.getElementById('apiMethodInput').value,
        path: document.getElementById('apiPathInput').value,
        description: document.getElementById('apiDescInput').value,
        category: document.getElementById('apiCategoryInput').value,
        source_database: (document.getElementById('apiSourceDatabaseInput').value || '').trim(),
        source_server: (document.getElementById('apiSourceServerInput').value || '').trim(),
        auth_required: document.getElementById('apiAuthInput').checked,
        rate_limit: parseInt(document.getElementById('apiRateLimitInput').value) || 0,
        cache_enabled: document.getElementById('apiCacheInput').checked,
        cache_ttl: parseInt(document.getElementById('apiCacheTTLInput').value) || 300,
        mock_response: mockResp,
        query_params: queryParams
    };

    postJSON('/api/v1/builder/apis', body).then(function(d) {
        if (d.status === 'success') {
            closeCreateAPIModal();
            loadBuilderSummary();
            loadCustomAPIs();
            addLog('Created API: ' + body.name, 'info');
        } else {
            alert(d.error || 'Failed to create API');
        }
    });
}

function loadBuilderDataSources() {
    fetchJSON('/api/admin/database/servers').then(function(d) {
        builderDataServers = d.servers || [];

        var dbSelect = document.getElementById('apiSourceDatabaseInput');
        if (!dbSelect) return;

        var existing = {};
        DEFAULT_BUILDER_DATABASES.forEach(function(db) { existing[db] = true; });
        (builderDataServers || []).forEach(function(s) {
            if (s && s.db_type) existing[s.db_type] = true;
        });

        var selectedDB = dbSelect.value;
        dbSelect.innerHTML = '<option value="">Select database...</option>';
        Object.keys(existing).sort().forEach(function(dbType) {
            dbSelect.innerHTML += '<option value="' + dbType + '">' + dbType.toUpperCase() + '</option>';
        });

        if (selectedDB && existing[selectedDB]) {
            dbSelect.value = selectedDB;
        }

        updateBuilderSourceServers();
    }).catch(function() {
        var dbSelect = document.getElementById('apiSourceDatabaseInput');
        if (!dbSelect) return;
        var selectedDB = dbSelect.value;
        dbSelect.innerHTML = '<option value="">Select database...</option>';
        DEFAULT_BUILDER_DATABASES.forEach(function(dbType) {
            dbSelect.innerHTML += '<option value="' + dbType + '">' + dbType.toUpperCase() + '</option>';
        });
        if (selectedDB) {
            dbSelect.value = selectedDB;
        }
        updateBuilderSourceServers();
    });
}

function updateBuilderSourceServers() {
    var dbTypeEl = document.getElementById('apiSourceDatabaseInput');
    var serverEl = document.getElementById('apiSourceServerInput');
    if (!dbTypeEl || !serverEl) return;

    var selectedDB = (dbTypeEl.value || '').trim();
    var previouslySelected = serverEl.value;
    var filtered = (builderDataServers || []).filter(function(s) {
        return !selectedDB || s.db_type === selectedDB;
    });

    serverEl.innerHTML = '<option value="">Default server for selected database</option>';
    filtered.forEach(function(s) {
        var name = (s.name || s.key || 'server') + (s.connected ? '' : ' (disconnected)');
        serverEl.innerHTML += '<option value="' + escapeHtml(s.key || '') + '">' + escapeHtml(name) + '</option>';
    });

    if (previouslySelected && filtered.some(function(s) { return s.key === previouslySelected; })) {
        serverEl.value = previouslySelected;
    }
}

function testCustomAPI(id) {
    postJSON('/api/v1/builder/apis/' + id + '/test', {}).then(function(d) {
        showResponse('API Test Result', 200, d, (d.method || 'GET') + ' ' + (d.path || ''));
        addLog('Tested API: ' + id, 'info');
    });
}

// ===================================================================
// CSV Upload & Dashboard Generation
// ===================================================================
function setupCSVDropZone() {
    var dz = document.getElementById('csvDropZone');
    if (!dz) return;
    dz.addEventListener('dragover', function(e) { e.preventDefault(); dz.classList.add('drag-over'); });
    dz.addEventListener('dragleave', function() { dz.classList.remove('drag-over'); });
    dz.addEventListener('drop', function(e) {
        e.preventDefault();
        dz.classList.remove('drag-over');
        if (e.dataTransfer.files.length > 0) {
            var fi = document.getElementById('csvFileInput');
            fi.files = e.dataTransfer.files;
            handleFileUpload();
        }
    });
}

function handleFileUpload() {
    var fileInput = document.getElementById('csvFileInput');
    if (!fileInput.files || !fileInput.files[0]) return;
    var file = fileInput.files[0];

    var formData = new FormData();
    formData.append('file', file);

    var token = localStorage.getItem('authToken');
    var headers = {};
    if (token) headers['Authorization'] = 'Bearer ' + token;

    fetch(BACKEND_URL + '/api/v1/builder/csv/upload', {
        method: 'POST', headers: headers, body: formData
    }).then(function(r) { return r.json(); }).then(function(d) {
        if (d.status === 'success') {
            currentCSVUploadId = d.upload.id;
            showCSVAnalysis(d.upload, d.can_convert_gis);
            var scanMsg = d.scan_safe ? ' (Security scan: PASSED)' : '';
            addLog('Uploaded file: ' + d.upload.filename + scanMsg, 'info');
            loadCSVHistory();
        } else {
            // Show scan failure details
            if (d.findings && d.findings.length > 0) {
                var msgs = d.findings.map(function(f) { return f.severity.toUpperCase() + ': ' + f.description; });
                alert('File rejected by security scanner:\\n\\n' + msgs.join('\\n'));
            } else {
                alert(d.error || 'Upload failed');
            }
        }
    }).catch(function(err) {
        alert('Upload error: ' + err.message);
    });
}

function showCSVAnalysis(upload, canGIS) {
    document.getElementById('csvAnalysisResult').style.display = 'block';
    document.getElementById('csvFilenameDisplay').textContent = upload.filename;

    var badges = '<span class="badge">' + upload.rows + ' rows</span> ' +
        '<span class="badge">' + upload.columns + ' cols</span> ' +
        '<span class="badge badge-' + (upload.has_geo_data ? 'success' : 'default') + '">' +
        (upload.has_geo_data ? 'Geo Data Detected' : 'No Geo Data') + '</span>';
    document.getElementById('csvAnalysisBadges').innerHTML = badges;

    // Column analysis
    var colHtml = '<div class="col-grid">';
    for (var i = 0; i < upload.column_names.length; i++) {
        var ct = (upload.column_types && upload.column_types[i]) || 'string';
        colHtml += '<div class="col-chip col-type-' + ct.replace(/[^a-z_]/g, '') + '">' +
            '<strong>' + escapeHtml(upload.column_names[i]) + '</strong><br><small>' + ct + '</small></div>';
    }
    colHtml += '</div>';
    document.getElementById('csvColumnAnalysis').innerHTML = colHtml;

    // Sample data table
    if (upload.sample_data && upload.sample_data.length > 0) {
        var tHtml = '<table class="admin-table compact"><thead><tr>';
        upload.column_names.forEach(function(c) { tHtml += '<th>' + escapeHtml(c) + '</th>'; });
        tHtml += '</tr></thead><tbody>';
        upload.sample_data.forEach(function(row) {
            tHtml += '<tr>';
            upload.column_names.forEach(function(c) { tHtml += '<td>' + escapeHtml(String(row[c] || '')) + '</td>'; });
            tHtml += '</tr>';
        });
        tHtml += '</tbody></table>';
        document.getElementById('csvSampleTable').innerHTML = tHtml;
    }

    var gisBtn = document.getElementById('btnGenerateGIS');
    if (gisBtn) gisBtn.style.display = canGIS ? 'inline-block' : 'none';
}

function generateCSVDashboard() {
    if (!currentCSVUploadId) return;
    postJSON('/api/v1/builder/csv/uploads/' + currentCSVUploadId + '/generate-dashboard', {}).then(function(d) {
        if (d.status === 'success') {
            addLog('Generated dashboard from CSV: ' + d.dashboard_id, 'info');
            alert('Dashboard created! Go to Analytics page to view it.');
            loadCSVHistory();
        } else {
            alert(d.error || 'Dashboard generation failed');
        }
    });
}

function generateCSVGIS() {
    if (!currentCSVUploadId) return;
    postJSON('/api/v1/builder/csv/uploads/' + currentCSVUploadId + '/generate-gis', {}).then(function(d) {
        if (d.status === 'success') {
            addLog('Generated GIS dataset from CSV: ' + d.dataset_id + ' (' + d.markers_created + ' markers)', 'info');
            alert('GIS dataset created with ' + d.markers_created + ' markers! Go to GIS page to view.');
            loadCSVHistory();
        } else {
            alert(d.error || 'GIS generation failed');
        }
    });
}

function loadCSVHistory() {
    fetchJSON('/api/v1/builder/csv/uploads').then(function(d) {
        var list = d.uploads || [];
        var el = document.getElementById('csvUploadHistory');
        if (!el) return;
        if (list.length === 0) {
            el.innerHTML = '<div class="empty-state">No CSV uploads yet</div>';
            return;
        }
        var html = '<table class="admin-table compact"><thead><tr><th>File</th><th>Type</th><th>Rows</th><th>Cols</th><th>Geo</th><th>Status</th><th>Dashboard</th><th>Actions</th></tr></thead><tbody>';
        list.forEach(function(u) {
            var safeId = u.id.replace(/'/g, "\\'");
            var dashActions = '';
            if (u.dashboard_id) {
                var safeDashId = u.dashboard_id.replace(/'/g, "\\'");
                dashActions = '<span>' + escapeHtml(u.dashboard_id) + '</span> <button class="btn-sm btn-del" onclick="deleteDashboard(\'' + safeDashId + '\')">Del</button>';
            } else if (u.gis_dashboard_id) {
                dashActions = escapeHtml(u.gis_dashboard_id);
            } else {
                dashActions = '-';
            }
            html += '<tr><td>' + escapeHtml(u.filename) + '</td><td>' + (u.file_type || 'csv').toUpperCase() + '</td><td>' + u.rows + '</td><td>' + u.columns + '</td>' +
                '<td>' + (u.has_geo_data ? 'Yes' : '-') + '</td>' +
                '<td><span class="status-badge status-' + u.status + '">' + u.status + '</span></td>' +
                '<td>' + dashActions + '</td>' +
                '<td><button class="btn-sm btn-del" onclick="deleteCSVUpload(\'' + safeId + '\')">Del</button></td></tr>';
        });
        html += '</tbody></table>';
        el.innerHTML = html;
    }).catch(function() {});
}

function deleteCSVUpload(id) {
    if (!confirm('Delete this CSV upload?')) return;
    deleteJSON('/api/v1/builder/csv/uploads/' + id).then(function() { loadCSVHistory(); });
}

// ===================================================================
// Dashboard <-> GIS Converter
// ===================================================================
function loadConverterDropdowns() {
    // Load analytics dashboards
    fetchJSON('/api/v1/analytics/dashboards').then(function(list) {
        var sel = document.getElementById('dashboardSelect');
        if (!sel) return;
        sel.innerHTML = '<option value="">Select a dashboard...</option>';
        (Array.isArray(list) ? list : []).forEach(function(d) {
            sel.innerHTML += '<option value="' + d.id + '">' + escapeHtml(d.name) + ' (' + (d.widgetCount || 0) + ' widgets)</option>';
        });
    }).catch(function() {});

    // Load GIS datasets
    fetchJSON('/api/v1/gis/datasets').then(function(d) {
        var datasets = d.datasets || d || [];
        var sel = document.getElementById('gisDatasetSelect');
        if (!sel) return;
        sel.innerHTML = '<option value="">Select a GIS dataset...</option>';
        (Array.isArray(datasets) ? datasets : []).forEach(function(ds) {
            sel.innerHTML += '<option value="' + ds.id + '">' + escapeHtml(ds.name || ds.id) + '</option>';
        });
    }).catch(function() {});
}

function analyzeDashToGIS() {
    var id = document.getElementById('dashboardSelect').value;
    var el = document.getElementById('dashToGISAnalysis');
    if (!id) { if (el) el.style.display = 'none'; return; }

    postJSON('/api/v1/builder/convert/analyze', { source_type: 'dashboard', source_id: id }).then(function(d) {
        el.style.display = 'block';
        currentDashMappings = d.field_mappings || [];
        var conf = Math.round((d.confidence || 0) * 100);
        var barColor = conf >= 50 ? '#10b981' : (conf >= 30 ? '#f59e0b' : '#ef4444');
        el.innerHTML =
            '<div class="conv-confidence"><div class="conf-bar" style="width:' + conf + '%;background:' + barColor + '"></div></div>' +
            '<div class="conv-text">Confidence: <strong>' + conf + '%</strong> &mdash; ' + escapeHtml(d.suggestion || '') + '</div>' +
            '<div class="conv-fields">Geo fields found: ' + (d.geo_fields_found || []).map(escapeHtml).join(', ') + '</div>';
        var btn = document.getElementById('btnConvertToGIS');
        if (btn) btn.style.display = d.can_convert ? 'inline-block' : 'none';
    }).catch(function() { el.style.display = 'none'; });
}

function convertDashboardToGIS() {
    var id = document.getElementById('dashboardSelect').value;
    if (!id) return;
    postJSON('/api/v1/builder/convert/dashboard-to-gis', { dashboard_id: id, field_mappings: currentDashMappings }).then(function(d) {
        if (d.status === 'success') {
            addLog('Converted dashboard to GIS: ' + d.dataset_id + ' (' + d.markers_created + ' markers)', 'info');
            alert(d.message || 'Conversion successful!');
            loadConversionHistory();
        } else { alert(d.error || 'Conversion failed'); }
    });
}

function analyzeGISToDash() {
    var id = document.getElementById('gisDatasetSelect').value;
    var el = document.getElementById('gisToDashAnalysis');
    if (!id) { if (el) el.style.display = 'none'; return; }

    postJSON('/api/v1/builder/convert/analyze', { source_type: 'gis', source_id: id }).then(function(d) {
        el.style.display = 'block';
        var conf = Math.round((d.confidence || 0) * 100);
        el.innerHTML =
            '<div class="conv-confidence"><div class="conf-bar" style="width:' + conf + '%;background:#10b981"></div></div>' +
            '<div class="conv-text">Confidence: <strong>' + conf + '%</strong> &mdash; ' + escapeHtml(d.suggestion || '') + '</div>';
        var btn = document.getElementById('btnConvertToDash');
        if (btn) btn.style.display = d.can_convert ? 'inline-block' : 'none';
    }).catch(function() { el.style.display = 'none'; });
}

function convertGISToDashboard() {
    var id = document.getElementById('gisDatasetSelect').value;
    if (!id) return;
    postJSON('/api/v1/builder/convert/gis-to-dashboard', { dataset_id: id }).then(function(d) {
        if (d.status === 'success') {
            addLog('Converted GIS to dashboard: ' + d.dashboard_id + ' (' + d.widget_count + ' widgets)', 'info');
            alert(d.message || 'Conversion successful!');
            loadConversionHistory();
        } else { alert(d.error || 'Conversion failed'); }
    });
}

function loadConversionHistory() {
    fetchJSON('/api/v1/builder/conversions').then(function(d) {
        var list = d.conversions || [];
        var el = document.getElementById('conversionHistoryList');
        if (!el) return;
        if (list.length === 0) {
            el.innerHTML = '<div class="empty-state">No conversions yet</div>';
            return;
        }
        var html = '<table class="admin-table compact"><thead><tr><th>Direction</th><th>Source</th><th>Target</th><th>Confidence</th><th>Status</th><th>Date</th></tr></thead><tbody>';
        list.forEach(function(c) {
            var dir = c.source_type === 'dashboard' ? 'Dashboard &rarr; GIS' : 'GIS &rarr; Dashboard';
            html += '<tr><td>' + dir + '</td><td>' + escapeHtml(c.source_id) + '</td><td>' + escapeHtml(c.target_id) + '</td>' +
                '<td>' + Math.round((c.confidence || 0) * 100) + '%</td>' +
                '<td><span class="status-badge status-' + c.status + '">' + c.status + '</span></td>' +
                '<td>' + new Date(c.created_at).toLocaleDateString() + '</td></tr>';
        });
        html += '</tbody></table>';
        el.innerHTML = html;
    }).catch(function() {});
}

// ===================================================================
// API Testing (original functionality)
// ===================================================================
function loadAPIs() {
    var apiCategories = {
        'Health & Status': [
            { method: 'GET', path: '/health', url: BACKEND_URL + '/health', description: 'Health check', auth: false },
            { method: 'GET', path: '/status', url: BACKEND_URL + '/status', description: 'Check all connections', auth: false },
        ],
        'Notifications': [
            { method: 'POST', path: '/api/notifications/send', url: BACKEND_URL + '/api/notifications/send', description: 'Send custom notification', auth: true, body: {title: 'Test', message: 'Test notification'} },
            { method: 'POST', path: '/api/notifications/health', url: BACKEND_URL + '/api/notifications/health', description: 'Send health notification', auth: true },
            { method: 'POST', path: '/api/notifications/status', url: BACKEND_URL + '/api/notifications/status', description: 'Send status notification', auth: true },
        ],
        'Admin - Database': [
            { method: 'GET', path: '/api/admin/database/list', url: BACKEND_URL + '/api/admin/database/list?db_type=mysql', description: 'List databases', auth: true },
            { method: 'POST', path: '/api/admin/database/create', url: BACKEND_URL + '/api/admin/database/create', description: 'Create database', auth: true, body: {database_name: 'test_db', db_type: 'mysql'} },
        ],
        'MySQL CRUD': [
            { method: 'GET', path: '/api/mysql/users', url: BACKEND_URL + '/api/mysql/users', description: 'List all users', auth: true },
            { method: 'POST', path: '/api/mysql/users', url: BACKEND_URL + '/api/mysql/users', description: 'Create user', auth: true, body: {name: 'John Doe', email: 'john@example.com'} },
        ],
        'PostgreSQL CRUD': [
            { method: 'GET', path: '/api/postgres/users', url: BACKEND_URL + '/api/postgres/users', description: 'List all users', auth: true },
            { method: 'POST', path: '/api/postgres/users', url: BACKEND_URL + '/api/postgres/users', description: 'Create user', auth: true, body: {name: 'Jane Doe', email: 'jane@example.com'} },
        ],
    };

    var html = '';
    for (var category in apiCategories) {
        var apis = apiCategories[category];
        html += '<div class="api-category"><div class="category-header">' + category + '</div><div class="api-items">';
        for (var i = 0; i < apis.length; i++) {
            var api = apis[i];
            html += '<button class="api-test-btn api-method-' + api.method.toLowerCase() + '" ' +
                'onclick="testAPI(\'' + api.method + '\', \'' + api.url + '\', ' +
                (api.body ? JSON.stringify(api.body).replace(/'/g, '&#39;') : 'null') + ', \'' + api.description + '\')">' +
                '<span class="api-method ' + api.method.toLowerCase() + '">' + api.method + '</span>' +
                '<span style="font-weight:500">' + api.path + '</span>' +
                '<span style="font-size:0.85em;color:#999">' + api.description + '</span>' +
                '</button>';
        }
        html += '</div></div>';
    }

    var el = document.getElementById('apiCategories');
    if (el) el.innerHTML = html;
}

function filterAPIs() {
    var searchTerm = document.getElementById('apiSearch').value.toLowerCase();
    var buttons = document.querySelectorAll('.api-test-btn');

    buttons.forEach(function(btn) {
        var text = btn.textContent.toLowerCase();
        var matches = text.indexOf(searchTerm) >= 0;
        var methodMatches = filteredMethod === 'ALL' || btn.classList.contains('api-method-' + filteredMethod.toLowerCase());
        btn.style.display = (matches && methodMatches) ? 'flex' : 'none';
    });
}

function filterByMethod(method) {
    filteredMethod = method;
    var filterBtns = document.querySelectorAll('.filter-btn');
    filterBtns.forEach(function(btn) { btn.classList.remove('active'); });
    if (event && event.target) event.target.classList.add('active');
    filterAPIs();
}

function testAPI(method, url, body, description) {
    var options = {
        method: method,
        headers: getAuthHeaders()
    };

    if (body && (method === 'POST' || method === 'PUT')) {
        options.body = typeof body === 'string' ? body : JSON.stringify(body);
    }

    showSpinner(description);

    fetch(url, options)
        .then(function(response) {
            var status = response.status;
            return response.json().catch(function() { return null; }).then(function(data) {
                return { status: status, data: data };
            });
        })
        .then(function(result) {
            showResponse(description, result.status, result.data, method + ' ' + url);
            addLog('API Call: ' + method + ' ' + url, 'info');
        })
        .catch(function(error) {
            showResponse(description, 'ERROR', {error: error.message}, method + ' ' + url);
            addLog('API Error: ' + error.message, 'error');
        });
}

// ===================================================================
// Modals & Logs
// ===================================================================
function showSpinner(title) {
    var modal = document.getElementById('responseModal');
    var body = document.getElementById('modalBody');
    document.getElementById('modalTitle').textContent = title;
    body.innerHTML = '<div class="spinner"></div><p class="loading">Loading...</p>';
    modal.style.display = 'flex';
}

function showResponse(title, status, data, endpoint) {
    var modal = document.getElementById('responseModal');
    var body = document.getElementById('modalBody');
    document.getElementById('modalTitle').textContent = title;

    var statusClass = 'success-message';
    if (typeof status === 'number' && status >= 400) {
        statusClass = 'error-message';
    } else if (typeof status === 'string' && status === 'ERROR') {
        statusClass = 'error-message';
    }

    var html = '<div class="' + statusClass + '"><strong>Status:</strong> ' + status + '</div>' +
        '<div style="margin-top:15px"><strong>Endpoint:</strong> <code>' + escapeHtml(String(endpoint)) + '</code></div>' +
        '<div style="margin-top:15px"><strong>Response:</strong></div>' +
        '<pre>' + escapeHtml(JSON.stringify(data, null, 2)) + '</pre>';

    body.innerHTML = html;
    modal.style.display = 'flex';
}

function closeResponseModal() {
    document.getElementById('responseModal').style.display = 'none';
}

function clearLogs() {
    document.getElementById('logsViewer').innerHTML = '';
}

function addLog(message, type) {
    var logViewer = document.getElementById('logsViewer');
    if (!logViewer) return;
    var timestamp = new Date().toLocaleTimeString();
    var entry = document.createElement('div');
    entry.className = 'log-entry ' + type;
    entry.innerHTML = '<span class="log-time">[' + timestamp + ']</span>' +
        '<span class="log-level">' + type.toUpperCase() + '</span>' +
        '<span class="log-message">' + escapeHtml(message) + '</span>';
    logViewer.insertBefore(entry, logViewer.firstChild);
}

// ===================================================================
// File Scanner (SafeGate Pipeline)
// ===================================================================
function setupScanDropZone() {
    var dz = document.getElementById('scanDropZone');
    if (!dz) return;
    dz.addEventListener('dragover', function(e) { e.preventDefault(); dz.classList.add('drag-over'); });
    dz.addEventListener('dragleave', function() { dz.classList.remove('drag-over'); });
    dz.addEventListener('drop', function(e) {
        e.preventDefault();
        dz.classList.remove('drag-over');
        if (e.dataTransfer.files.length > 0) {
            var fi = document.getElementById('scanFileInput');
            fi.files = e.dataTransfer.files;
            handleFileScan();
        }
    });
}

function handleFileScan() {
    var fileInput = document.getElementById('scanFileInput');
    if (!fileInput.files || !fileInput.files[0]) return;
    var file = fileInput.files[0];

    var formData = new FormData();
    formData.append('file', file);

    var token = localStorage.getItem('authToken');
    var headers = {};
    if (token) headers['Authorization'] = 'Bearer ' + token;

    document.getElementById('scanResult').style.display = 'block';
    document.getElementById('scanFilenameDisplay').textContent = 'Scanning ' + file.name + '...';
    document.getElementById('scanBadges').innerHTML = '<span class="badge">Scanning...</span>';
    document.getElementById('scanFindings').innerHTML = '<div class="loading">Running SafeGate pipeline...</div>';

    fetch(BACKEND_URL + '/api/v1/builder/scanner/scan', {
        method: 'POST', headers: headers, body: formData
    }).then(function(r) { return r.json(); }).then(function(d) {
        if (d.status === 'success') {
            showScanResult(d.scan);
            addLog('Scanned file: ' + d.scan.filename + ' — ' + (d.safe ? 'SAFE' : 'THREATS FOUND'), d.safe ? 'info' : 'error');
            loadScanHistory();
        } else {
            document.getElementById('scanFindings').innerHTML = '<div class="error-state">' + escapeHtml(d.error || 'Scan failed') + '</div>';
        }
    }).catch(function(err) {
        document.getElementById('scanFindings').innerHTML = '<div class="error-state">Scan error: ' + escapeHtml(err.message) + '</div>';
    });
}

function showScanResult(scan) {
    document.getElementById('scanFilenameDisplay').textContent = scan.filename;
    var safeBadge = scan.safe
        ? '<span class="badge badge-success">✅ SAFE</span>'
        : '<span class="badge badge-danger">⚠️ THREATS FOUND</span>';
    document.getElementById('scanBadges').innerHTML =
        safeBadge + ' ' +
        '<span class="badge">' + formatFileSize(scan.file_size) + '</span> ' +
        '<span class="badge">' + scan.findings.length + ' finding(s)</span>';

    var html = '';
    if (scan.findings.length === 0) {
        html = '<div class="empty-state" style="padding:10px">No security issues found.</div>';
    } else {
        html = '<table class="admin-table compact"><thead><tr><th>Severity</th><th>Scanner</th><th>Description</th><th>Details</th></tr></thead><tbody>';
        scan.findings.forEach(function(f) {
            var sevClass = f.severity === 'critical' || f.severity === 'high' ? 'status-inactive' :
                           f.severity === 'medium' ? 'status-draft' : 'status-active';
            html += '<tr><td><span class="status-badge ' + sevClass + '">' + f.severity.toUpperCase() + '</span></td>' +
                '<td>' + escapeHtml(f.scanner) + '</td>' +
                '<td>' + escapeHtml(f.description) + '</td>' +
                '<td><small>' + escapeHtml(f.details || '') + '</small></td></tr>';
        });
        html += '</tbody></table>';
    }
    document.getElementById('scanFindings').innerHTML = html;
}

function formatFileSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

function loadScannerHealth() {
    fetchJSON('/api/v1/builder/scanner/health').then(function(d) {
        var el = document.getElementById('scannerHealth');
        if (!el) return;
        var html = '<div class="col-grid">';
        (d.scanners || []).forEach(function(name) {
            html += '<div class="col-chip col-type-string"><strong>' + escapeHtml(name) + '</strong><br><small>Active</small></div>';
        });
        html += '</div><p style="margin-top:8px;color:#888">Total scans performed: ' + (d.total_scans || 0) + '</p>';
        el.innerHTML = html;
    }).catch(function() {
        var el = document.getElementById('scannerHealth');
        if (el) el.innerHTML = '<div class="empty-state">Could not load scanner status</div>';
    });
}

function loadScanHistory() {
    fetchJSON('/api/v1/builder/scanner/scans').then(function(d) {
        var list = d.scans || [];
        var el = document.getElementById('scanHistory');
        if (!el) return;
        if (list.length === 0) {
            el.innerHTML = '<div class="empty-state">No scans yet</div>';
            return;
        }
        var html = '<table class="admin-table compact"><thead><tr><th>File</th><th>Size</th><th>Result</th><th>Findings</th><th>Date</th></tr></thead><tbody>';
        list.forEach(function(s) {
            var resultBadge = s.safe
                ? '<span class="status-badge status-active">Safe</span>'
                : '<span class="status-badge status-inactive">Unsafe</span>';
            html += '<tr><td>' + escapeHtml(s.filename) + '</td>' +
                '<td>' + formatFileSize(s.file_size) + '</td>' +
                '<td>' + resultBadge + '</td>' +
                '<td>' + s.findings.length + '</td>' +
                '<td>' + new Date(s.scanned_at).toLocaleString() + '</td></tr>';
        });
        html += '</tbody></table>';
        el.innerHTML = html;
    }).catch(function() {});
}

// Close modals on outside click
window.onclick = function(event) {
    var responseModal = document.getElementById('responseModal');
    if (event.target === responseModal) responseModal.style.display = 'none';
    var createModal = document.getElementById('createAPIModal');
    if (event.target === createModal) createModal.style.display = 'none';
};

// ===================================================================
// Role-Based Access Control
// ===================================================================
function canModify() {
    var role = localStorage.getItem('userRole') || 'user';
    return role === 'admin' || role === 'manager';
}

function applyRoleRestrictions() {
    var role = localStorage.getItem('userRole') || 'user';
    if (role === 'admin' || role === 'manager') return; // full access

    // Normal users: hide create/delete/modify controls
    var hideSelectors = [
        '.btn-create-api',         // Create API button
        '#createAPIBtn',           // Create API button by id
        '.btn-del',                // Delete buttons in API table
        '.btn-danger',             // Danger buttons
    ];

    // Hide "Create API" button 
    var createBtns = document.querySelectorAll('button');
    createBtns.forEach(function(btn) {
        var text = btn.textContent.trim().toLowerCase();
        if (text.includes('create api') || text.includes('+ create') || text.includes('delete')) {
            btn.style.display = 'none';
        }
    });

    // Add a read-only banner
    var container = document.querySelector('.admin-container') || document.querySelector('.manager-header');
    if (container) {
        var banner = document.createElement('div');
        banner.style.cssText = 'background:rgba(59,130,246,0.15);color:#3b82f6;padding:10px 20px;border-radius:8px;margin-bottom:16px;text-align:center;font-weight:600;';
        banner.textContent = '👁️ View-Only Mode — Your role (' + role + ') allows viewing and calling APIs only';
        container.insertBefore(banner, container.firstChild);
    }
}

// Override delete/create functions for role checks
function deleteCustomAPI(id) {
    if (!canModify()) { alert('You do not have permission to delete APIs. Contact an admin or manager.'); return; }
    if (!confirm('Delete this API?')) return;
    deleteJSON('/api/v1/builder/apis/' + id).then(function(d) {
        if (d.status === 'success' || d.status === 'ok') { addLog('Deleted API: ' + id, 'warn'); loadBuilderSummary(); loadCustomAPIs(); }
        else { alert(d.error || 'Failed to delete API'); }
    });
}

function deleteDashboard(dashId) {
    if (!canModify()) { alert('You do not have permission to delete dashboards. Contact an admin or manager.'); return; }
    if (!confirm('Delete this dashboard? This cannot be undone.')) return;
    deleteJSON('/api/v1/builder/dashboards/' + dashId).then(function(d) {
        if (d.status === 'success') { addLog('Deleted dashboard: ' + dashId, 'warn'); loadCSVHistory(); }
        else { alert(d.error || 'Failed to delete dashboard'); }
    });
}

function openCreateAPIModal() {
    if (!canModify()) { alert('You do not have permission to create APIs. Contact an admin or manager.'); return; }
    loadBuilderDataSources();
    document.getElementById('createAPIModal').style.display = 'flex';
}
