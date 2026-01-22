// Admin Dashboard JS
const BACKEND_URL = (() => {
    const elem = document.getElementById('backendURL');
    if (elem && elem.textContent) {
        return elem.textContent.trim();
    }
    return 'http://localhost:8000';
})();

let filteredMethod = 'ALL';

console.log('Admin - Backend URL:', BACKEND_URL);

window.addEventListener('DOMContentLoaded', function() {
    loadAPIs();
});

function switchTab(tabName) {
    // Hide all tabs
    const tabs = document.querySelectorAll('.tab-content');
    tabs.forEach(function(tab) { tab.classList.remove('active'); });
    
    // Remove active from buttons
    const buttons = document.querySelectorAll('.tab-btn');
    buttons.forEach(function(btn) { btn.classList.remove('active'); });
    
    // Show selected tab
    const selectedTab = document.getElementById(tabName);
    if (selectedTab) selectedTab.classList.add('active');
    
    // Add active to clicked button
    event.target.classList.add('active');
}

function loadAPIs() {
    const apiCategories = {
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

    let html = '';
    for (const [category, apis] of Object.entries(apiCategories)) {
        html += '<div class="api-category">' +
            '<div class="category-header">' + category + '</div>' +
            '<div class="api-items">';
        
        for (let i = 0; i < apis.length; i++) {
            const api = apis[i];
            html += '<button class="api-test-btn api-method-' + api.method.toLowerCase() + '" ' +
                'onclick="testAPI(\'' + api.method + '\', \'' + api.url + '\', ' + 
                (api.body ? JSON.stringify(api.body).replace(/'/g, '&#39;') : 'null') + ', \'' + api.description + '\')">' +
                '<span class="api-method ' + api.method.toLowerCase() + '">' + api.method + '</span>' +
                '<span style="font-weight: 500;">' + api.path + '</span>' +
                '<span style="font-size: 0.85em; color: #999;">' + api.description + '</span>' +
                '</button>';
        }
        
        html += '</div></div>';
    }
    
    document.getElementById('apiCategories').innerHTML = html;
}

function filterAPIs() {
    const searchTerm = document.getElementById('apiSearch').value.toLowerCase();
    const buttons = document.querySelectorAll('.api-test-btn');
    
    buttons.forEach(function(btn) {
        const text = btn.textContent.toLowerCase();
        const matches = text.includes(searchTerm);
        const methodMatches = filteredMethod === 'ALL' || btn.classList.contains('api-method-' + filteredMethod.toLowerCase());
        
        btn.style.display = (matches && methodMatches) ? 'flex' : 'none';
    });
}

function filterByMethod(method) {
    filteredMethod = method;
    
    const filterBtns = document.querySelectorAll('.filter-btn');
    filterBtns.forEach(function(btn) { btn.classList.remove('active'); });
    event.target.classList.add('active');
    
    filterAPIs();
}

function testAPI(method, url, body, description) {
    const options = {
        method: method,
        headers: getAuthHeaders()
    };

    if (body && (method === 'POST' || method === 'PUT')) {
        options.body = typeof body === 'string' ? body : JSON.stringify(body);
    }

    showSpinner(description);

    fetch(url, options)
        .then(function(response) {
            const status = response.status;
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

function showSpinner(title) {
    const modal = document.getElementById('responseModal');
    const body = document.getElementById('modalBody');
    document.getElementById('modalTitle').textContent = title;
    body.innerHTML = '<div class="spinner"></div><p class="loading">Loading...</p>';
    modal.style.display = 'flex';
}

function showResponse(title, status, data, endpoint) {
    const modal = document.getElementById('responseModal');
    const body = document.getElementById('modalBody');
    document.getElementById('modalTitle').textContent = title;
    
    let statusClass = 'success-message';
    if (typeof status === 'number' && status >= 400) {
        statusClass = 'error-message';
    } else if (typeof status === 'string' && status === 'ERROR') {
        statusClass = 'error-message';
    }
    
    let html = '<div class="' + statusClass + '"><strong>Status:</strong> ' + status + '</div>' +
        '<div style="margin-top: 15px;"><strong>Endpoint:</strong> <code>' + endpoint + '</code></div>' +
        '<div style="margin-top: 15px;"><strong>Response:</strong></div>' +
        '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
    
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
    const logViewer = document.getElementById('logsViewer');
    const timestamp = new Date().toLocaleTimeString();
    const entry = document.createElement('div');
    entry.className = 'log-entry ' + type;
    entry.innerHTML = '<span class="log-time">[' + timestamp + ']</span>' +
        '<span class="log-level">' + type.toUpperCase() + '</span>' +
        '<span class="log-message">' + message + '</span>';
    logViewer.insertBefore(entry, logViewer.firstChild);
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modal = document.getElementById('responseModal');
    if (event.target === modal) {
        modal.style.display = 'none';
    }
};
