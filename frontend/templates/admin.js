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
let customAPIById = {};
let graphQLAPIById = {};
let graphQLFormMode = 'create';
let editingGraphQLApiId = '';

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
    loadGraphQLBuilderSummary();
    loadGraphQLCustomAPIs();
    loadBuilderDataSources();
    loadCSVHistory();
    loadAPIs();
    setupCSVDropZone();
    setupScanDropZone();
    loadScannerHealth();
    applyRoleRestrictions();
    initAdminCertificateActions();
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
    if (tabName === 'graphql-api-builder') { loadGraphQLBuilderSummary(); loadGraphQLCustomAPIs(); }
    if (tabName === 'csv-upload') { loadCSVHistory(); }
    if (tabName === 'converter') { loadConverterDropdowns(); loadConversionHistory(); }
    if (tabName === 'file-scanner') { loadScanHistory(); loadScannerHealth(); }
    if (tabName === 'api-testing') { loadAPIs(); }
    if (tabName === 'graphql-studio') { loadAdminGraphQLSchemaInfo(); }
    if (tabName === 'control-plane') { refreshAdminControlPlaneData(); }
    if (tabName === 'settings') { loadAdminCertificatePanel(); }
}

// ===================================================================
// API Builder
// ===================================================================
function loadBuilderSummary() {
    fetchJSON('/api/v1/builder/summary?api_type=rest').then(function(d) {
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
    var q = '/api/v1/builder/apis?api_type=rest&';
    if (cat) q += 'category=' + encodeURIComponent(cat) + '&';
    if (status) q += 'status=' + encodeURIComponent(status);

    fetchJSON(q).then(function(d) {
        var list = d.apis || [];
        customAPIById = {};
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
            customAPIById[api.id] = api;
            var safeId = api.id.replace(/'/g, "\\'");
            var actionsHtml = '<button class="btn-sm btn-test" onclick="testCustomAPI(\'' + safeId + '\')">Test</button> ';
            if (canModify()) {
                var isActive = (api.status || '').toLowerCase() === 'active';
                actionsHtml += '<button class="btn-sm btn-edit" onclick="openEditCustomAPI(\'' + safeId + '\')">Edit</button> ';
                actionsHtml += '<button class="btn-sm btn-toggle" onclick="toggleCustomAPIStatus(\'' + safeId + '\')">' + (isActive ? 'Deactivate' : 'Activate') + '</button> ';
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

function countSQLPlaceholders(query) {
    var m = String(query || '').match(/\?/g);
    return m ? m.length : 0;
}

function getSQLTemplateFromAPI(api) {
    if (api && api.sql_template) return String(api.sql_template).trim();
    if (api && api.mock_response && typeof api.mock_response === 'object' && api.mock_response.query) {
        return String(api.mock_response.query).trim();
    }
    return '';
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

    var sqlTemplate = (document.getElementById('apiSQLTemplateInput').value || '').trim();
    var sqlPlaceholderCount = countSQLPlaceholders(sqlTemplate);

    var body = {
        api_type: 'rest',
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
        sql_template: sqlTemplate,
        mock_response: mockResp,
        query_params: queryParams
    };

    if (body.source_database && !body.sql_template) {
        alert('SQL Template is required when Source Database is selected.');
        return;
    }
    if (sqlPlaceholderCount > 0 && queryParams.length === 0) {
        alert('Define Query Parameters for each SQL placeholder (?); one parameter per placeholder in order.');
        return;
    }

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

function loadBuilderDataSources(onDone) {
    fetchJSON('/api/admin/database/servers').then(function(d) {
        builderDataServers = d.servers || [];

        var dbSelect = document.getElementById('apiSourceDatabaseInput');
        var gqlDbSelect = document.getElementById('gqlSourceDatabaseInput');
        var editDbSelect = document.getElementById('editApiSourceDatabaseInput');
        if (!dbSelect && !gqlDbSelect && !editDbSelect) {
            if (typeof onDone === 'function') onDone();
            return;
        }

        var existing = {};
        DEFAULT_BUILDER_DATABASES.forEach(function(db) { existing[db] = true; });
        (builderDataServers || []).forEach(function(s) {
            if (s && s.db_type) existing[s.db_type] = true;
        });

        function populateDatabaseSelect(selectEl) {
            if (!selectEl) return;
            var selectedDB = selectEl.value;
            selectEl.innerHTML = '<option value="">Select database...</option>';
            Object.keys(existing).sort().forEach(function(dbType) {
                selectEl.innerHTML += '<option value="' + dbType + '">' + dbType.toUpperCase() + '</option>';
            });
            if (selectedDB && existing[selectedDB]) {
                selectEl.value = selectedDB;
            }
        }

        populateDatabaseSelect(dbSelect);
        populateDatabaseSelect(gqlDbSelect);
        populateDatabaseSelect(editDbSelect);

        updateBuilderSourceServers();
        updateGraphQLBuilderSourceServers();
        updateEditBuilderSourceServers();
        if (typeof onDone === 'function') onDone();
    }).catch(function() {
        var dbSelect = document.getElementById('apiSourceDatabaseInput');
        var gqlDbSelect = document.getElementById('gqlSourceDatabaseInput');
        var editDbSelect = document.getElementById('editApiSourceDatabaseInput');
        if (!dbSelect && !gqlDbSelect && !editDbSelect) {
            if (typeof onDone === 'function') onDone();
            return;
        }

        function populateDefault(selectEl) {
            if (!selectEl) return;
            var selectedDB = selectEl.value;
            selectEl.innerHTML = '<option value="">Select database...</option>';
            DEFAULT_BUILDER_DATABASES.forEach(function(dbType) {
                selectEl.innerHTML += '<option value="' + dbType + '">' + dbType.toUpperCase() + '</option>';
            });
            if (selectedDB) {
                selectEl.value = selectedDB;
            }
        }

        populateDefault(dbSelect);
        populateDefault(gqlDbSelect);
        populateDefault(editDbSelect);

        updateBuilderSourceServers();
        updateGraphQLBuilderSourceServers();
        updateEditBuilderSourceServers();
        if (typeof onDone === 'function') onDone();
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

function updateEditBuilderSourceServers() {
    var dbTypeEl = document.getElementById('editApiSourceDatabaseInput');
    var serverEl = document.getElementById('editApiSourceServerInput');
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

function loadGraphQLBuilderSummary() {
    fetchJSON('/api/v1/builder/summary?api_type=graphql').then(function(d) {
        var el = document.getElementById('graphqlBuilderSummary');
        if (!el) return;
        el.innerHTML =
            '<div class="summary-card"><div class="sc-value">' + (d.total_apis || 0) + '</div><div class="sc-label">GraphQL APIs</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.active || 0) + '</div><div class="sc-label">Active</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.draft || 0) + '</div><div class="sc-label">Draft</div></div>' +
            '<div class="summary-card"><div class="sc-value">' + (d.total_hits || 0) + '</div><div class="sc-label">Total Hits</div></div>';
    }).catch(function() {});
}

function loadGraphQLCustomAPIs() {
    var catEl = document.getElementById('graphqlApiCategoryFilter');
    var statEl = document.getElementById('graphqlApiStatusFilter');
    var cat = catEl ? catEl.value : '';
    var status = statEl ? statEl.value : '';
    var q = '/api/v1/builder/apis?api_type=graphql&';
    if (cat) q += 'category=' + encodeURIComponent(cat) + '&';
    if (status) q += 'status=' + encodeURIComponent(status);

    fetchJSON(q).then(function(d) {
        var list = d.apis || [];
        graphQLAPIById = {};
        var el = document.getElementById('graphqlApiBuilderList');
        if (!el) return;
        if (list.length === 0) {
            el.innerHTML = '<div class="empty-state">No GraphQL APIs found. Click <strong>+ Create GraphQL API</strong> to get started.</div>';
            return;
        }

        var html = '<table class="admin-table"><thead><tr>' +
            '<th>Name</th><th>Operation</th><th>Endpoint</th><th>Category</th><th>Source DB</th><th>Source Server</th><th>Status</th><th>Hits</th><th>Actions</th>' +
            '</tr></thead><tbody>';

        list.forEach(function(api) {
            graphQLAPIById[api.id] = api;
            var safeId = api.id.replace(/'/g, "\\'");
            var actionsHtml = '<button class="btn-sm btn-test" onclick="testGraphQLCustomAPI(\'' + safeId + '\')">Test</button> ';
            if (canModify()) {
                var isActive = (api.status || '').toLowerCase() === 'active';
                actionsHtml += '<button class="btn-sm btn-edit" onclick="openEditGraphQLCustomAPI(\'' + safeId + '\')">Edit</button> ';
                actionsHtml += '<button class="btn-sm btn-toggle" onclick="toggleGraphQLCustomAPIStatus(\'' + safeId + '\')">' + (isActive ? 'Deactivate' : 'Activate') + '</button> ';
                actionsHtml += '<button class="btn-sm btn-del" onclick="deleteGraphQLCustomAPI(\'' + safeId + '\')">Del</button>';
            }
            html += '<tr>' +
                '<td>' + escapeHtml(api.name) + '</td>' +
                '<td>' + escapeHtml(api.graphql_operation_name || '-') + '</td>' +
                '<td><code>' + escapeHtml(api.path || '/api/graphql') + '</code></td>' +
                '<td>' + escapeHtml(api.category || '-') + '</td>' +
                '<td>' + escapeHtml(api.source_database || '-') + '</td>' +
                '<td>' + escapeHtml(api.source_server || 'default') + '</td>' +
                '<td><span class="status-badge status-' + api.status + '">' + api.status + '</span></td>' +
                '<td>' + (api.hit_count || 0) + '</td>' +
                '<td>' + actionsHtml + '</td>' +
                '</tr>';
        });

        html += '</tbody></table>';
        el.innerHTML = html;
    }).catch(function() {
        var el = document.getElementById('graphqlApiBuilderList');
        if (el) el.innerHTML = '<div class="error-state">Failed to load GraphQL APIs</div>';
    });
}

function resetGraphQLFormMode() {
    graphQLFormMode = 'create';
    editingGraphQLApiId = '';

    var titleEl = document.querySelector('#createGraphQLAPIModal .modal-header h2');
    if (titleEl) titleEl.textContent = '🧩 Create GraphQL API';

    var submitBtn = document.querySelector('#createGraphQLAPIForm button[type="submit"]');
    if (submitBtn) submitBtn.textContent = 'Create GraphQL API';
}

function updateGraphQLBuilderSourceServers() {
    var dbTypeEl = document.getElementById('gqlSourceDatabaseInput');
    var serverEl = document.getElementById('gqlSourceServerInput');
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

function openCreateGraphQLAPIModal() {
    if (!canModify()) { alert('You do not have permission to create APIs. Contact an admin, manager, or system-manager.'); return; }
    resetGraphQLFormMode();
    loadBuilderDataSources();
    document.getElementById('createGraphQLAPIModal').style.display = 'flex';
}

function closeCreateGraphQLAPIModal() {
    resetGraphQLFormMode();
    document.getElementById('createGraphQLAPIModal').style.display = 'none';
    document.getElementById('createGraphQLAPIForm').reset();
    var ttlGroup = document.getElementById('gqlCacheTTLGroup');
    if (ttlGroup) ttlGroup.style.display = 'none';
    var pathEl = document.getElementById('gqlPathInput');
    if (pathEl && !pathEl.value) {
        pathEl.value = '/api/graphql';
    }
    updateGraphQLBuilderSourceServers();
}

function toggleGraphQLCacheTTL() {
    var checked = document.getElementById('gqlCacheInput').checked;
    var group = document.getElementById('gqlCacheTTLGroup');
    if (group) group.style.display = checked ? 'block' : 'none';
}

function parseBuilderParams(raw) {
    var params = [];
    if (!raw) return params;
    raw.split('\n').forEach(function(line) {
        var parts = line.trim().split(':');
        if (parts.length >= 2) {
            params.push({ name: parts[0].trim(), type: parts[1].trim(), required: parts.length > 2 && parts[2].trim() === 'true' });
        }
    });
    return params;
}

function submitCreateGraphQLAPI(e) {
    e.preventDefault();

    var mockRaw = document.getElementById('gqlMockResponseInput').value.trim();
    var mockResp = null;
    if (mockRaw) {
        try { mockResp = JSON.parse(mockRaw); } catch (err) { alert('Invalid JSON in mock response'); return; }
    }

    var body = {
        name: document.getElementById('gqlApiNameInput').value,
        method: 'POST',
        path: (document.getElementById('gqlPathInput').value || '/api/graphql').trim(),
        graphql_query: document.getElementById('gqlQueryInput').value,
        graphql_operation_name: (document.getElementById('gqlOperationNameInput').value || '').trim(),
        description: document.getElementById('gqlDescInput').value,
        category: document.getElementById('gqlApiCategoryInput').value,
        source_database: (document.getElementById('gqlSourceDatabaseInput').value || '').trim(),
        source_server: (document.getElementById('gqlSourceServerInput').value || '').trim(),
        auth_required: document.getElementById('gqlAuthInput').checked,
        rate_limit: parseInt(document.getElementById('gqlRateLimitInput').value, 10) || 0,
        cache_enabled: document.getElementById('gqlCacheInput').checked,
        cache_ttl: parseInt(document.getElementById('gqlCacheTTLInput').value, 10) || 300,
        mock_response: mockResp,
        query_params: parseBuilderParams(document.getElementById('gqlQueryParamsInput').value.trim())
    };

    var isEditMode = graphQLFormMode === 'edit' && !!editingGraphQLApiId;
    var req;
    if (isEditMode) {
        body.api_type = 'graphql';
        req = putJSON('/api/v1/builder/apis/' + encodeURIComponent(editingGraphQLApiId), body);
    } else {
        body.api_type = 'graphql';
        req = postJSON('/api/v1/builder/apis', body);
    }

    req.then(function(d) {
        if (d.status === 'success') {
            closeCreateGraphQLAPIModal();
            loadGraphQLBuilderSummary();
            loadGraphQLCustomAPIs();
            if (isEditMode) {
                addLog('Updated GraphQL API: ' + body.name, 'info');
            } else {
                addLog('Created GraphQL API: ' + body.name, 'info');
            }
        } else {
            alert(d.error || 'Failed to save GraphQL API');
        }
    });
}

function testGraphQLCustomAPI(id) {
    postJSON('/api/v1/builder/apis/' + id + '/test', {}).then(function(d) {
        showResponse('GraphQL API Test Result', 200, d, 'POST ' + (d.path || '/api/graphql'));
        addLog('Tested GraphQL API: ' + id, 'info');
    });
}

function deleteGraphQLCustomAPI(id) {
    if (!canModify()) { alert('You do not have permission to delete APIs. Contact an admin, manager, or system-manager.'); return; }
    var api = graphQLAPIById[id] || {};
    var apiName = api.name || id;
    if (!confirmDeleteAPI('GraphQL API', apiName)) return;
    deleteJSON('/api/v1/builder/apis/' + id).then(function(d) {
        if (d.status === 'success' || d.status === 'ok') {
            addLog('Deleted GraphQL API: ' + id, 'warn');
            loadGraphQLBuilderSummary();
            loadGraphQLCustomAPIs();
        } else {
            alert(d.error || 'Failed to delete GraphQL API');
        }
    });
}

function openEditGraphQLCustomAPI(id) {
    if (!canModify()) { alert('You do not have permission to edit APIs. Contact an admin, manager, or system-manager.'); return; }
    var api = graphQLAPIById[id];
    if (!api) { alert('GraphQL API details not found. Please refresh and try again.'); return; }

    graphQLFormMode = 'edit';
    editingGraphQLApiId = id;

    var titleEl = document.querySelector('#createGraphQLAPIModal .modal-header h2');
    if (titleEl) titleEl.textContent = '✏️ Edit GraphQL API';
    var submitBtn = document.querySelector('#createGraphQLAPIForm button[type="submit"]');
    if (submitBtn) submitBtn.textContent = 'Save Changes';

    loadBuilderDataSources(function() {
        document.getElementById('gqlApiNameInput').value = api.name || '';
        document.getElementById('gqlApiCategoryInput').value = api.category || 'custom';
        document.getElementById('gqlOperationNameInput').value = api.graphql_operation_name || '';
        document.getElementById('gqlPathInput').value = (api.path || '/api/graphql').trim();
        document.getElementById('gqlSourceDatabaseInput').value = api.source_database || '';
        updateGraphQLBuilderSourceServers();
        document.getElementById('gqlSourceServerInput').value = api.source_server || '';
        document.getElementById('gqlDescInput').value = api.description || '';
        document.getElementById('gqlQueryInput').value = api.graphql_query || '';
        document.getElementById('gqlAuthInput').checked = !!api.auth_required;
        document.getElementById('gqlRateLimitInput').value = api.rate_limit || 0;
        document.getElementById('gqlCacheInput').checked = !!api.cache_enabled;
        document.getElementById('gqlCacheTTLInput').value = api.cache_ttl || 300;
        toggleGraphQLCacheTTL();
        document.getElementById('gqlMockResponseInput').value = api.mock_response ? JSON.stringify(api.mock_response, null, 2) : '';

        var qp = Array.isArray(api.query_params) ? api.query_params : [];
        var qpLines = qp.map(function(p) {
            var pName = (p && p.name) ? p.name : '';
            var pType = (p && p.type) ? p.type : 'string';
            var pReq = (p && p.required) ? 'true' : 'false';
            return pName + ':' + pType + ':' + pReq;
        });
        document.getElementById('gqlQueryParamsInput').value = qpLines.join('\n');

        document.getElementById('createGraphQLAPIModal').style.display = 'flex';
    });
}

function toggleGraphQLCustomAPIStatus(id) {
    if (!canModify()) { alert('You do not have permission to update API status. Contact an admin, manager, or system-manager.'); return; }

    var api = graphQLAPIById[id];
    if (!api) { alert('GraphQL API details not found. Please refresh and try again.'); return; }

    var currentStatus = String(api.status || '').toLowerCase();
    var nextStatus = currentStatus === 'active' ? 'inactive' : 'active';

    putJSON('/api/v1/builder/apis/' + encodeURIComponent(id), { status: nextStatus }).then(function(d) {
        if (d.status === 'success' || d.status === 'ok') {
            addLog('Updated GraphQL API status: ' + id + ' -> ' + nextStatus, 'info');
            loadGraphQLBuilderSummary();
            loadGraphQLCustomAPIs();
        } else {
            alert(d.error || 'Failed to update GraphQL API status');
        }
    }).catch(function() {
        alert('Failed to update GraphQL API status');
    });
}

function testCustomAPI(id) {
    postJSON('/api/v1/builder/apis/' + id + '/test', {}).then(function(d) {
        showResponse('API Test Result', 200, d, (d.method || 'GET') + ' ' + (d.path || ''));
        addLog('Tested API: ' + id, 'info');
    });
}

function openEditCustomAPI(id) {
    if (!canModify()) { alert('You do not have permission to edit APIs. Contact an admin, manager, or system-manager.'); return; }
    var api = customAPIById[id];
    if (!api) { alert('API details not found. Please refresh and try again.'); return; }

    loadBuilderDataSources(function() {
        document.getElementById('editApiIdInput').value = api.id || '';
        document.getElementById('editApiNameInput').value = api.name || '';
        document.getElementById('editApiMethodInput').value = api.method || 'GET';
        document.getElementById('editApiPathInput').value = api.path || '';
        document.getElementById('editApiCategoryInput').value = api.category || 'custom';
        document.getElementById('editApiSourceDatabaseInput').value = api.source_database || '';
        updateEditBuilderSourceServers();
        document.getElementById('editApiSourceServerInput').value = api.source_server || '';
        document.getElementById('editApiDescInput').value = api.description || '';
        document.getElementById('editApiSQLTemplateInput').value = getSQLTemplateFromAPI(api);
        document.getElementById('editApiAuthInput').checked = !!api.auth_required;
        document.getElementById('editApiCacheInput').checked = !!api.cache_enabled;
        document.getElementById('editApiCacheTTLInput').value = api.cache_ttl || 300;
        toggleEditCacheTTL();
        document.getElementById('editApiRateLimitInput').value = api.rate_limit || 0;
        document.getElementById('editApiStatusInput').value = api.status || 'active';
        document.getElementById('editAPIModal').style.display = 'flex';
    });
}

function closeEditAPIModal() {
    document.getElementById('editAPIModal').style.display = 'none';
    document.getElementById('editAPIForm').reset();
    var ttlGroup = document.getElementById('editCacheTTLGroup');
    if (ttlGroup) ttlGroup.style.display = 'none';
}

function toggleEditCacheTTL() {
    var checked = document.getElementById('editApiCacheInput').checked;
    var group = document.getElementById('editCacheTTLGroup');
    if (group) group.style.display = checked ? 'block' : 'none';
}

function submitEditAPI(e) {
    e.preventDefault();
    if (!canModify()) { alert('You do not have permission to edit APIs. Contact an admin, manager, or system-manager.'); return; }

    var id = document.getElementById('editApiIdInput').value;
    if (!id) {
        alert('Invalid API ID');
        return;
    }

    var body = {
        name: document.getElementById('editApiNameInput').value,
        method: document.getElementById('editApiMethodInput').value,
        path: document.getElementById('editApiPathInput').value,
        description: document.getElementById('editApiDescInput').value,
        sql_template: (document.getElementById('editApiSQLTemplateInput').value || '').trim(),
        category: document.getElementById('editApiCategoryInput').value,
        source_database: (document.getElementById('editApiSourceDatabaseInput').value || '').trim(),
        source_server: (document.getElementById('editApiSourceServerInput').value || '').trim(),
        auth_required: document.getElementById('editApiAuthInput').checked,
        cache_enabled: document.getElementById('editApiCacheInput').checked,
        cache_ttl: parseInt(document.getElementById('editApiCacheTTLInput').value, 10) || 300,
        rate_limit: parseInt(document.getElementById('editApiRateLimitInput').value, 10) || 0,
        status: document.getElementById('editApiStatusInput').value
    };

    if (body.source_database && !body.sql_template) {
        alert('SQL Template is required when Source Database is selected.');
        return;
    }

    putJSON('/api/v1/builder/apis/' + encodeURIComponent(id), body).then(function(d) {
        if (d.status === 'success' || d.status === 'ok') {
            closeEditAPIModal();
            loadBuilderSummary();
            loadCustomAPIs();
            addLog('Updated API: ' + id, 'info');
        } else {
            alert(d.error || 'Failed to update API');
        }
    }).catch(function() {
        alert('Failed to update API');
    });
}

function toggleCustomAPIStatus(id) {
    if (!canModify()) { alert('You do not have permission to update API status. Contact an admin, manager, or system-manager.'); return; }

    var api = customAPIById[id];
    if (!api) { alert('API details not found. Please refresh and try again.'); return; }

    var currentStatus = String(api.status || '').toLowerCase();
    var nextStatus = currentStatus === 'active' ? 'inactive' : 'active';

    putJSON('/api/v1/builder/apis/' + encodeURIComponent(id), { status: nextStatus }).then(function(d) {
        if (d.status === 'success' || d.status === 'ok') {
            addLog('Updated API status: ' + id + ' -> ' + nextStatus, 'info');
            loadBuilderSummary();
            loadCustomAPIs();
        } else {
            alert(d.error || 'Failed to update API status');
        }
    }).catch(function() {
        alert('Failed to update API status');
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
    var el = document.getElementById('apiCategories');
    if (el) el.innerHTML = '<div class="loading">Loading custom runtime endpoints...</div>';

    Promise.all([
        fetchJSON('/api/v1/builder/apis?api_type=rest&status=active').catch(function() { return { apis: [] }; }),
        fetchJSON('/api/v1/builder/apis?api_type=graphql&status=active').catch(function() { return { apis: [] }; })
    ]).then(function(results) {
        var restAPIs = (results[0] && results[0].apis) ? results[0].apis : [];
        var graphqlAPIs = (results[1] && results[1].apis) ? results[1].apis : [];

        var categories = {
            'Core Platform': [
                { method: 'GET', path: '/health', url: BACKEND_URL + '/health', description: 'Health check', body: null },
                { method: 'GET', path: '/status', url: BACKEND_URL + '/status', description: 'Platform status', body: null }
            ],
            'Notifications': [
                {
                    method: 'POST',
                    path: '/api/v1/notifications/send',
                    url: BACKEND_URL + '/api/v1/notifications/send',
                    description: 'Send custom notification to Discord',
                    body: {
                        title: 'AxiomNizam Notification',
                        message: 'Notification API restored and working.',
                        type: 'info'
                    }
                },
                {
                    method: 'POST',
                    path: '/api/v1/notifications/health',
                    url: BACKEND_URL + '/api/v1/notifications/health',
                    description: 'Send health check notification',
                    body: null
                },
                {
                    method: 'POST',
                    path: '/api/v1/notifications/status',
                    url: BACKEND_URL + '/api/v1/notifications/status',
                    description: 'Send platform status notification',
                    body: null
                },
                {
                    method: 'GET',
                    path: '/api/v1/notifications/status',
                    url: BACKEND_URL + '/api/v1/notifications/status',
                    description: 'Get notification service status',
                    body: null
                }
            ],
            'Custom REST Runtime (API Builder)': restAPIs.map(function(api) {
                var method = String(api.method || 'GET').toUpperCase();
                var runtimePath = normalizeRuntimePath(api.path || '/');
                var url = BACKEND_URL + '/api/custom' + (runtimePath === '/' ? '' : runtimePath);
                var body = null;
                if (method !== 'GET') {
                    var paramsBody = {};
                    (api.query_params || []).forEach(function(p) {
                        paramsBody[p.name] = sampleParamValue(p);
                    });
                    body = { params: paramsBody };
                } else if ((api.query_params || []).length > 0) {
                    var queryPairs = [];
                    api.query_params.forEach(function(p) {
                        queryPairs.push(encodeURIComponent(p.name) + '=' + encodeURIComponent(sampleParamValue(p)));
                    });
                    if (queryPairs.length > 0) {
                        url += (url.indexOf('?') >= 0 ? '&' : '?') + queryPairs.join('&');
                    }
                }

                return {
                    method: method,
                    path: '/api/custom' + (runtimePath === '/' ? '' : runtimePath),
                    url: url,
                    description: api.name || 'Custom runtime API',
                    body: body
                };
            }),
            'Custom GraphQL Runtime (API Builder)': graphqlAPIs.map(function(api) {
                return {
                    method: String(api.method || 'POST').toUpperCase(),
                    path: api.path || '/api/graphql',
                    url: BACKEND_URL + normalizeRuntimePath(api.path || '/api/graphql'),
                    description: api.name || 'Custom GraphQL API',
                    body: {
                        query: api.graphql_query || 'query { __typename }',
                        operationName: api.graphql_operation_name || '',
                        variables: {}
                    }
                };
            })
        };

        var html = '';
        for (var category in categories) {
            var apis = categories[category] || [];
            if (apis.length === 0) continue;
            html += '<div class="api-category"><div class="category-header">' + escapeHtml(category) + '</div><div class="api-items">';
            for (var i = 0; i < apis.length; i++) {
                var api = apis[i];
                html += '<button class="api-test-btn api-method-' + String(api.method || 'GET').toLowerCase() + '" ' +
                    'onclick="testAPI(\'' + api.method + '\', \'' + api.url + '\', ' +
                    (api.body ? JSON.stringify(api.body).replace(/'/g, '&#39;') : 'null') + ', \'' + escapeHtml(api.description).replace(/'/g, '&#39;') + '\')">' +
                    '<span class="api-method ' + String(api.method || 'GET').toLowerCase() + '">' + escapeHtml(String(api.method || 'GET')) + '</span>' +
                    '<span style="font-weight:500">' + escapeHtml(api.path) + '</span>' +
                    '<span style="font-size:0.85em;color:#999">' + escapeHtml(api.description) + '</span>' +
                    '</button>';
            }
            html += '</div></div>';
        }

        if (!html) {
            html = '<div class="empty-state">No custom APIs available yet. Create APIs in API Builder and return here to test runtime endpoints.</div>';
        }

        if (el) el.innerHTML = html;
        filterAPIs();
    }).catch(function() {
        if (el) el.innerHTML = '<div class="error-state">Failed to load runtime endpoints</div>';
    });
}

function normalizeRuntimePath(path) {
    var normalized = '/' + String(path || '').trim().replace(/^\/+|\/+$/g, '');
    if (normalized === '/api/custom') return '/';
    if (normalized.indexOf('/api/custom/') === 0) {
        normalized = normalized.replace('/api/custom', '');
        normalized = '/' + normalized.trim().replace(/^\/+|\/+$/g, '');
    }
    if (normalized === '/api/graphql') return '/api/graphql';
    return normalized;
}

function sampleParamValue(param) {
    if (!param) return 'sample';
    if (param.default && String(param.default).trim() !== '') return param.default;
    var t = String(param.type || '').toLowerCase();
    if (t === 'int' || t === 'integer' || t === 'number' || t === 'float' || t === 'decimal') return 1;
    if (t === 'bool' || t === 'boolean') return true;
    return 'sample';
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

    if (body && (method === 'POST' || method === 'PUT' || method === 'PATCH')) {
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
// GraphQL Studio + Control Plane (Admin)
// ===================================================================
function adminApiCall(method, path, body) {
    var options = {
        method: method,
        headers: getAuthHeaders()
    };
    if (body !== undefined && body !== null) {
        options.body = JSON.stringify(body);
    }

    return fetch(BACKEND_URL + path, options).then(function(response) {
        return response.text().then(function(text) {
            var parsed;
            try {
                parsed = text ? JSON.parse(text) : {};
            } catch (e) {
                parsed = { raw: text };
            }
            if (!response.ok) {
                var msg = (parsed && (parsed.error || parsed.message)) || ('Request failed with status ' + response.status);
                var err = new Error(msg);
                err.status = response.status;
                err.response = parsed;
                throw err;
            }
            return { status: response.status, data: parsed };
        });
    });
}

function parseJSONInput(elementId, fallback) {
    var el = document.getElementById(elementId);
    if (!el) return fallback;
    var raw = (el.value || '').trim();
    if (!raw) return fallback;
    return JSON.parse(raw);
}

function setAdminControlPlaneOutput(title, payload) {
    var el = document.getElementById('adminControlPlaneOutput');
    if (!el) return;
    el.textContent = title + '\n\n' + JSON.stringify(payload, null, 2);
}

function getAdminControlPlaneInput() {
    var namespace = (document.getElementById('adminCpNamespace').value || 'default').trim() || 'default';
    var kind = (document.getElementById('adminCpKind').value || 'apis').trim().toLowerCase();
    var name = (document.getElementById('adminCpName').value || '').trim();
    return { namespace: namespace, kind: kind, name: name };
}

function canControlPlaneWrite() {
    var role = (localStorage.getItem('userRole') || 'user').toLowerCase();
    return role === 'admin' || role === 'system-manager';
}

function ensureControlPlaneWrite(actionLabel) {
    if (canControlPlaneWrite()) return true;
    alert('RBAC: only admin or system-manager can ' + actionLabel + '.');
    return false;
}

function runAdminGraphQLQuery() {
    var queryEl = document.getElementById('adminGraphQLQuery');
    var opEl = document.getElementById('adminGraphQLOperation');
    var resultEl = document.getElementById('adminGraphQLResult');
    if (!queryEl || !resultEl) return;

    var query = (queryEl.value || '').trim();
    if (!query) {
        alert('GraphQL query is required.');
        return;
    }

    var variables;
    try {
        variables = parseJSONInput('adminGraphQLVariables', {});
    } catch (e) {
        alert('Invalid JSON in GraphQL variables: ' + e.message);
        return;
    }

    resultEl.textContent = 'Running query...';
    adminApiCall('POST', '/api/graphql', {
        query: query,
        variables: variables,
        operationName: (opEl && opEl.value ? opEl.value.trim() : '') || undefined
    }).then(function(result) {
        resultEl.textContent = JSON.stringify(result.data, null, 2);
        addLog('GraphQL query executed', 'info');
    }).catch(function(err) {
        resultEl.textContent = JSON.stringify({
            error: err.message,
            status: err.status || 'n/a',
            details: err.response || {}
        }, null, 2);
        addLog('GraphQL query failed: ' + err.message, 'error');
    });
}

function loadAdminGraphQLSchemaInfo() {
    var resultEl = document.getElementById('adminGraphQLResult');
    if (!resultEl) return;
    adminApiCall('GET', '/api/graphql/schema').then(function(result) {
        resultEl.textContent = JSON.stringify(result.data, null, 2);
    }).catch(function(err) {
        resultEl.textContent = JSON.stringify({
            error: err.message,
            details: err.response || {}
        }, null, 2);
    });
}

function adminApplyResource() {
    if (!ensureControlPlaneWrite('apply resources')) return;
    var meta = getAdminControlPlaneInput();
    var body;
    try {
        body = parseJSONInput('adminCpBody', {});
    } catch (e) {
        alert('Invalid JSON in resource body: ' + e.message);
        return;
    }

    if (!body.metadata) body.metadata = {};
    if (!body.metadata.name && meta.name) {
        body.metadata.name = meta.name;
    }

    adminApiCall('POST', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind), body)
        .then(function(result) {
            setAdminControlPlaneOutput('Resource applied', result.data);
            addLog('Applied resource ' + (body.metadata.name || ''), 'info');
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Apply failed', { error: err.message, details: err.response || {} });
        });
}

function adminListResources() {
    var meta = getAdminControlPlaneInput();
    adminApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind))
        .then(function(result) {
            setAdminControlPlaneOutput('Resource list', result.data);
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('List failed', { error: err.message, details: err.response || {} });
        });
}

function adminGetResource() {
    var meta = getAdminControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    adminApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name))
        .then(function(result) {
            setAdminControlPlaneOutput('Resource detail', result.data);
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Get failed', { error: err.message, details: err.response || {} });
        });
}

function adminGetResourceStatus() {
    var meta = getAdminControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    adminApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name) + '/status')
        .then(function(result) {
            setAdminControlPlaneOutput('Resource status', result.data);
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Status failed', { error: err.message, details: err.response || {} });
        });
}

function adminGetResourceEvents() {
    var meta = getAdminControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    adminApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name) + '/events')
        .then(function(result) {
            setAdminControlPlaneOutput('Resource events', result.data);
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Events failed', { error: err.message, details: err.response || {} });
        });
}

function adminDeleteResource() {
    if (!ensureControlPlaneWrite('delete resources')) return;
    var meta = getAdminControlPlaneInput();
    if (!meta.name) { alert('Resource name is required.'); return; }
    adminApiCall('DELETE', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind) + '/' + encodeURIComponent(meta.name))
        .then(function(result) {
            setAdminControlPlaneOutput('Resource deleted', result.data);
            addLog('Deleted resource ' + meta.name, 'warn');
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Delete failed', { error: err.message, details: err.response || {} });
        });
}

function adminRunWorkflow() {
    if (!ensureControlPlaneWrite('run workflows')) return;
    var name = (document.getElementById('adminWorkflowName').value || '').trim();
    if (!name) { alert('Workflow name is required.'); return; }
    adminApiCall('POST', '/api/v1/workflows/' + encodeURIComponent(name) + '/run', {})
        .then(function(result) {
            setAdminControlPlaneOutput('Workflow run requested', result.data);
        })
        .catch(function(err) {
            setAdminControlPlaneOutput('Workflow run failed', { error: err.message, details: err.response || {} });
        });
}

function adminListDatasources() {
    adminApiCall('GET', '/api/v1/datasources').then(function(result) {
        setAdminControlPlaneOutput('Datasources', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('List datasources failed', { error: err.message, details: err.response || {} });
    });
}

function adminGetDatasource() {
    var name = (document.getElementById('adminDatasourceName').value || '').trim();
    if (!name) { alert('Datasource name is required.'); return; }
    adminApiCall('GET', '/api/v1/datasources/' + encodeURIComponent(name)).then(function(result) {
        setAdminControlPlaneOutput('Datasource detail', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Get datasource failed', { error: err.message, details: err.response || {} });
    });
}

function adminTestDatasource() {
    if (!ensureControlPlaneWrite('test datasources')) return;
    var name = (document.getElementById('adminDatasourceName').value || '').trim();
    if (!name) { alert('Datasource name is required.'); return; }
    adminApiCall('POST', '/api/v1/datasources/' + encodeURIComponent(name) + '/test', {}).then(function(result) {
        setAdminControlPlaneOutput('Datasource test result', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Test datasource failed', { error: err.message, details: err.response || {} });
    });
}

function adminListJobs() {
    adminApiCall('GET', '/api/v1/jobs').then(function(result) {
        setAdminControlPlaneOutput('Jobs', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('List jobs failed', { error: err.message, details: err.response || {} });
    });
}

function adminGetJob() {
    var id = (document.getElementById('adminJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    adminApiCall('GET', '/api/v1/jobs/' + encodeURIComponent(id)).then(function(result) {
        setAdminControlPlaneOutput('Job detail', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Get job failed', { error: err.message, details: err.response || {} });
    });
}

function adminRunJob() {
    if (!ensureControlPlaneWrite('run jobs')) return;
    var id = (document.getElementById('adminJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    adminApiCall('POST', '/api/v1/jobs/' + encodeURIComponent(id) + '/run', {}).then(function(result) {
        setAdminControlPlaneOutput('Job run requested', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Run job failed', { error: err.message, details: err.response || {} });
    });
}

function adminGetJobLogs() {
    var id = (document.getElementById('adminJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    adminApiCall('GET', '/api/v1/jobs/' + encodeURIComponent(id) + '/logs').then(function(result) {
        setAdminControlPlaneOutput('Job logs', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Get job logs failed', { error: err.message, details: err.response || {} });
    });
}

function adminCancelJob() {
    if (!ensureControlPlaneWrite('cancel jobs')) return;
    var id = (document.getElementById('adminJobId').value || '').trim();
    if (!id) { alert('Job ID/name is required.'); return; }
    adminApiCall('POST', '/api/v1/jobs/' + encodeURIComponent(id) + '/cancel', {}).then(function(result) {
        setAdminControlPlaneOutput('Job cancel requested', result.data);
    }).catch(function(err) {
        setAdminControlPlaneOutput('Cancel job failed', { error: err.message, details: err.response || {} });
    });
}

function refreshAdminControlPlaneData() {
    var meta = getAdminControlPlaneInput();
    Promise.all([
        adminApiCall('GET', '/api/v1/namespaces/' + encodeURIComponent(meta.namespace) + '/' + encodeURIComponent(meta.kind)),
        adminApiCall('GET', '/api/v1/datasources'),
        adminApiCall('GET', '/api/v1/jobs')
    ]).then(function(results) {
        setAdminControlPlaneOutput('Control plane snapshot', {
            resources: results[0].data,
            datasources: results[1].data,
            jobs: results[2].data
        });
    }).catch(function(err) {
        setAdminControlPlaneOutput('Refresh failed', { error: err.message, details: err.response || {} });
    });
}

function getAdminCertName() {
    var el = document.getElementById('adminCertName');
    if (!el) return '';
    return (el.value || '').trim();
}

function setAdminCertificateOutput(title, payload) {
    var el = document.getElementById('adminCertificateOutput');
    if (!el) return;
    el.textContent = title + '\n\n' + JSON.stringify(payload, null, 2);
}

function loadAdminCertificatePanel() {
    var el = document.getElementById('adminCertificateOutput');
    if (el && !el.textContent.trim()) {
        el.textContent = 'Kubernetes certificate status will appear here.';
    }
}

function initAdminCertificateActions() {
    var checkBtn = document.getElementById('adminCertCheckBtn');
    var dryRunBtn = document.getElementById('adminCertDryRunBtn');
    var renewBtn = document.getElementById('adminCertRenewBtn');

    if (checkBtn && !checkBtn.dataset.bound) {
        checkBtn.addEventListener('click', function() { adminCheckCertificateExpiry(); });
        checkBtn.dataset.bound = '1';
    }
    if (dryRunBtn && !dryRunBtn.dataset.bound) {
        dryRunBtn.addEventListener('click', function() { adminRenewCertificate(true); });
        dryRunBtn.dataset.bound = '1';
    }
    if (renewBtn && !renewBtn.dataset.bound) {
        renewBtn.addEventListener('click', function() { adminRenewCertificate(false); });
        renewBtn.dataset.bound = '1';
    }
}

function adminCheckCertificateExpiry() {
    if (!ensureControlPlaneWrite('view certificate status')) return;

    var cert = getAdminCertName();
    var path = '/api/admin/certificates/status';
    if (cert) {
        path += '?cert=' + encodeURIComponent(cert);
    }

    adminApiCall('GET', path).then(function(result) {
        setAdminCertificateOutput('Certificate status', result.data);
        addLog('Checked certificate expiry' + (cert ? ' for ' + cert : ''), 'info');
    }).catch(function(err) {
        setAdminCertificateOutput('Certificate status failed', { error: err.message, details: err.response || {} });
    });
}

function adminRenewCertificate(dryRun) {
    if (!ensureControlPlaneWrite('renew certificates')) return;

    var cert = getAdminCertName();
    var body = { dry_run: !!dryRun };
    if (cert) body.cert = cert;

    adminApiCall('POST', '/api/admin/certificates/renew', body).then(function(result) {
        setAdminCertificateOutput(dryRun ? 'Certificate renew dry run' : 'Certificate renew result', result.data);
        addLog((dryRun ? 'Prepared' : 'Triggered') + ' certificate renewal' + (cert ? ' for ' + cert : ''), dryRun ? 'info' : 'warn');
    }).catch(function(err) {
        setAdminCertificateOutput('Certificate renew failed', { error: err.message, details: err.response || {} });
    });
}

window.adminCheckCertificateExpiry = adminCheckCertificateExpiry;
window.adminRenewCertificate = adminRenewCertificate;

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
    var editModal = document.getElementById('editAPIModal');
    if (event.target === editModal) editModal.style.display = 'none';
    var createGraphQLModal = document.getElementById('createGraphQLAPIModal');
    if (event.target === createGraphQLModal) closeCreateGraphQLAPIModal();
};

// ===================================================================
// Role-Based Access Control
// ===================================================================
function canModify() {
    var role = localStorage.getItem('userRole') || 'user';
    return role === 'admin' || role === 'manager' || role === 'system-manager';
}

function applyRoleRestrictions() {
    var role = localStorage.getItem('userRole') || 'user';
    if (role === 'admin' || role === 'manager' || role === 'system-manager') return; // full access

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

function confirmDeleteAPI(kindLabel, apiName) {
    var label = kindLabel || 'API';
    var name = String(apiName || '').trim() || 'unnamed-api';
    if (!confirm('Delete ' + label + ' "' + name + '"? This action cannot be undone.')) {
        return false;
    }

    var ack = prompt('Type DELETE to confirm deleting "' + name + '"');
    if (ack === null) {
        return false;
    }
    if (String(ack).trim().toUpperCase() !== 'DELETE') {
        alert('Delete cancelled: confirmation text did not match DELETE.');
        return false;
    }
    return true;
}

// Override delete/create functions for role checks
function deleteCustomAPI(id) {
    if (!canModify()) { alert('You do not have permission to delete APIs. Contact an admin or manager.'); return; }
    var api = customAPIById[id] || {};
    var apiName = api.name || id;
    if (!confirmDeleteAPI('REST API', apiName)) return;
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
    if (!canModify()) { alert('You do not have permission to create APIs. Contact an admin, manager, or system-manager.'); return; }
    loadBuilderDataSources();
    document.getElementById('createAPIModal').style.display = 'flex';
}
