// Admin Dashboard JS — API Builder, CSV-to-Dashboard, Dashboard↔GIS Converter
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
    loadCSVHistory();
    loadAPIs();
    setupCSVDropZone();
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
            '<th>Method</th><th>Name</th><th>Path</th><th>Category</th><th>Status</th><th>Hits</th><th>Actions</th>' +
            '</tr></thead><tbody>';
        list.forEach(function(api) {
            var safeId = api.id.replace(/'/g, "\\'");
            html += '<tr>' +
                '<td><span class="method-badge method-' + api.method.toLowerCase() + '">' + api.method + '</span></td>' +
                '<td>' + escapeHtml(api.name) + '</td>' +
                '<td><code>' + escapeHtml(api.path) + '</code></td>' +
                '<td>' + escapeHtml(api.category || '-') + '</td>' +
                '<td><span class="status-badge status-' + api.status + '">' + api.status + '</span></td>' +
                '<td>' + (api.hit_count || 0) + '</td>' +
                '<td>' +
                    '<button class="btn-sm btn-test" onclick="testCustomAPI(\'' + safeId + '\')">Test</button> ' +
                    '<button class="btn-sm btn-del" onclick="deleteCustomAPI(\'' + safeId + '\')">Del</button>' +
                '</td></tr>';
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

function openCreateAPIModal() {
    document.getElementById('createAPIModal').style.display = 'flex';
}
function closeCreateAPIModal() {
    document.getElementById('createAPIModal').style.display = 'none';
    document.getElementById('createAPIForm').reset();
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
        auth_required: document.getElementById('apiAuthInput').checked,
        rate_limit: parseInt(document.getElementById('apiRateLimitInput').value) || 0,
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

function testCustomAPI(id) {
    postJSON('/api/v1/builder/apis/' + id + '/test', {}).then(function(d) {
        showResponse('API Test Result', 200, d, (d.method || 'GET') + ' ' + (d.path || ''));
        addLog('Tested API: ' + id, 'info');
    });
}

function deleteCustomAPI(id) {
    if (!confirm('Delete this API?')) return;
    deleteJSON('/api/v1/builder/apis/' + id).then(function() {
        loadBuilderSummary();
        loadCustomAPIs();
        addLog('Deleted API: ' + id, 'warn');
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
            handleCSVUpload();
        }
    });
}

function handleCSVUpload() {
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
            addLog('Uploaded CSV: ' + d.upload.filename, 'info');
            loadCSVHistory();
        } else {
            alert(d.error || 'Upload failed');
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
        var html = '<table class="admin-table compact"><thead><tr><th>File</th><th>Rows</th><th>Cols</th><th>Geo</th><th>Status</th><th>Dashboard</th><th>Actions</th></tr></thead><tbody>';
        list.forEach(function(u) {
            var safeId = u.id.replace(/'/g, "\\'");
            html += '<tr><td>' + escapeHtml(u.filename) + '</td><td>' + u.rows + '</td><td>' + u.columns + '</td>' +
                '<td>' + (u.has_geo_data ? 'Yes' : '-') + '</td>' +
                '<td><span class="status-badge status-' + u.status + '">' + u.status + '</span></td>' +
                '<td>' + (u.dashboard_id || u.gis_dashboard_id || '-') + '</td>' +
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

// Close modals on outside click
window.onclick = function(event) {
    var responseModal = document.getElementById('responseModal');
    if (event.target === responseModal) responseModal.style.display = 'none';
    var createModal = document.getElementById('createAPIModal');
    if (event.target === createModal) createModal.style.display = 'none';
};
